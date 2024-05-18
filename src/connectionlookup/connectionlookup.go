package connectionlookup

import (
	"maps"
	"sync"

	"github.com/lxzan/gws"
)

type Connection struct {
	Id     string
	Socket *gws.Conn
	// Reference the key/val lists this connection is in. Only used
	// so this connection knows what key/val pairs it has.
	KeyVals     map[int]*ConnectionLockList
	KeyValsLock sync.RWMutex

	JsonExtractVars   *map[string]string
	StreamIncludeTags *[]string
}

func NewConnection(id string, socket *gws.Conn) *Connection {
	return &Connection{
		Id:                id,
		Socket:            socket,
		KeyVals:           make(map[int]*ConnectionLockList),
		JsonExtractVars:   nil,
		StreamIncludeTags: nil,
	}
}

type ConnectionList map[string]*Connection
type ConnectionLockList struct {
	ConnectionList
	Key    string
	KeyVal string
	Lock   sync.RWMutex
}

type ConnectionLookup struct {
	strings   *StringMap
	redisSync *RedisSync

	// Fast lookup of ID > connection
	connections     map[string]*Connection
	connectionsLock sync.RWMutex

	/*
		map[the key][the val][con id] = connection
		Root >
			Key1 >
				Val1 >
					connectionId: connection
					connectionId: connection
				Val2 >
					connectionId: connection
			Key2 >
				Val1 >
					connectionId: connection
					connectionId: connection
				Val2 >
					connectionId: connection
	*/
	tree     map[string]map[string]*ConnectionLockList
	treeLock sync.RWMutex
}

func NewConnectionLookup(redisAddr string) (*ConnectionLookup, error) {
	var sync *RedisSync
	if redisAddr != "" {
		s, err := NewRedisSync(redisAddr)
		if err != nil {
			return nil, err
		}
		sync = s
	}

	lookup := &ConnectionLookup{
		strings:     NewStringMap(),
		connections: make(map[string]*Connection),
		tree:        make(map[string]map[string]*ConnectionLockList),
		redisSync:   sync,
	}

	return lookup, nil
}

func (c *ConnectionLookup) SetKeys(con *Connection, keys map[string]string) {
	newKeys := map[string]string{}
	delKeys := []string{}

	for key, keyVal := range keys {
		// Keep track of vals to update redis after
		if keyVal != "" {
			newKeys[key] = keyVal
		} else {
			delKeys = append(delKeys, key)
		}

		c.treeLock.Lock()
		keyList, isOk := c.tree[key]
		if !isOk {
			if keyVal == "" {
				// Empty value means were deleting it. So don't add a new entry
				continue
			}

			keyList = make(map[string]*ConnectionLockList)
			c.tree[key] = keyList
		}
		c.treeLock.Unlock()

		// Empty value = removing it
		if keyVal == "" {
			// We don't know the existing value yet so check every value we know of.
			// High cardinality may slow this down though.
			// TODO: Perhaps keep track of key=val pairs on the Connection instance too? Maybe not
			//       due to memory usage though. Or a pointer to the ConnectionLockList...
			for checkingVal, checkingValList := range keyList {
				_, conExists := checkingValList.ConnectionList[con.Id]
				if !conExists {
					continue
				}

				checkingValList.Lock.Lock()
				delete(checkingValList.ConnectionList, con.Id)
				checkingValList.Lock.Unlock()

				// Cleanup empty lists
				if len(checkingValList.ConnectionList) == 0 {
					c.treeLock.Lock() // < TODO: Why this lock treelock?
					delete(keyList, checkingVal)
					c.treeLock.Unlock()
				}

				if len(c.tree[key]) == 0 {
					c.treeLock.Lock()
					delete(c.tree, key)
					c.treeLock.Unlock()
				}
			}

			con.KeyValsLock.Lock()
			delete(con.KeyVals, c.strings.Get(key))
			con.KeyValsLock.Unlock()

			continue
		}

		valList, valListExist := keyList[keyVal]
		if !valListExist {
			valList = &ConnectionLockList{
				Key:            key,
				KeyVal:         keyVal,
				ConnectionList: make(ConnectionList),
			}
			keyList[keyVal] = valList
		}

		valList.Lock.Lock()
		valList.ConnectionList[con.Id] = con
		valList.Lock.Unlock()

		// Let the connection itself know what key/val pairs it has
		con.KeyValsLock.Lock()
		con.KeyVals[c.strings.Get(key)] = valList
		con.KeyValsLock.Unlock()
	}

	if c.redisSync != nil && len(newKeys) > 0 {
		c.redisSync.UpdateConnectionTags(con, newKeys)
	}
	if c.redisSync != nil && len(delKeys) > 0 {
		c.redisSync.RemoveConnectionTags(con, delKeys)
	}
}

func (c *ConnectionLookup) AddConnection(con *Connection, keys map[string]string) {
	c.connectionsLock.Lock()
	c.connections[con.Id] = con
	c.connectionsLock.Unlock()

	if c.redisSync != nil {
		c.redisSync.AddConnection(con)
	}

	c.SetKeys(con, keys)
}

func (c *ConnectionLookup) RemoveConnection(con *Connection) {
	c.connectionsLock.Lock()
	delete(c.connections, con.Id)
	c.connectionsLock.Unlock()

	con.KeyValsLock.Lock()
	defer con.KeyValsLock.Unlock()

	c.treeLock.Lock()
	for _, keyval := range con.KeyVals {
		keyval.Lock.Lock()
		delete(keyval.ConnectionList, con.Id)

		// Cleanup any empty lists
		if len(keyval.ConnectionList) == 0 {
			delete(c.tree[keyval.Key], keyval.KeyVal)
		}

		// Cleanup any empty lists
		if len(c.tree[keyval.Key]) == 0 {
			delete(c.tree, keyval.Key)
		}

		keyval.Lock.Unlock()
	}
	c.treeLock.Unlock()

	for k := range con.KeyVals {
		delete(con.KeyVals, k)
	}

	if c.redisSync != nil {
		c.redisSync.RemoveConnection(con)
	}
}

func (c *ConnectionLookup) GetConnectionById(id string) (*Connection, bool) {
	c.connectionsLock.RLock()
	con, isOk := c.connections[id]
	c.connectionsLock.RUnlock()
	return con, isOk
}

// Find all connections that have all the given keys
func (c *ConnectionLookup) GetConnectionsWithKeys(keys map[string]string) []*Connection {
	// Keep track of how many key=val combos were matched against a connection. If a connection
	// has len(keys) matches at the end, then we know it matched all the keys
	connections := make(map[*Connection]int)

	matches := make([]*Connection, 0)

	c.treeLock.RLock()
	defer c.treeLock.RUnlock()

	for key, keyVal := range keys {
		listOfVals, isOk := c.tree[key]
		if !isOk {
			// If we don't have a list for this key then we definitely don't have all keys.
			return []*Connection{}
		}

		listOfCons, isOk := listOfVals[keyVal]
		if !isOk {
			// No connection has this key=val combo so no point continuing
			return []*Connection{}
		}

		listOfCons.Lock.RLock()
		for _, con := range listOfCons.ConnectionList {
			connections[con]++
			if connections[con] == len(keys) {
				matches = append(matches, con)
			}
		}
		listOfCons.Lock.RUnlock()
	}

	return matches
}

func (c *ConnectionLookup) GetAllKeys() []string {
	keys := []string{}

	c.treeLock.RLock()
	for key := range c.tree {
		keys = append(keys, key)
	}
	c.treeLock.RUnlock()

	return keys
}

func (c *ConnectionLookup) GetAllKeysAndValue() map[string][]string {
	keys := map[string][]string{}

	c.treeLock.RLock()
	for key, valList := range c.tree {
		keys[key] = make([]string, len(valList))
		cnt := 0
		for val := range valList {
			keys[key][cnt] = val
			cnt++
		}
	}
	c.treeLock.RUnlock()

	return keys
}

func (c *ConnectionLookup) NumConnections() int {
	c.connectionsLock.RLock()
	defer c.connectionsLock.RUnlock()
	return len(c.connections)
}

func (c *ConnectionLookup) GetConnections() map[string]*Connection {
	c.connectionsLock.RLock()
	defer c.connectionsLock.RUnlock()
	return maps.Clone(c.connections)
}

func (c *ConnectionLookup) DumpConnections() []map[string]string {
	entries := make([]map[string]string, 0)

	c.connectionsLock.RLock()
	defer c.connectionsLock.RUnlock()

	c.treeLock.RLock()
	defer c.treeLock.RUnlock()

	for _, con := range c.connections {
		entry := map[string]string{
			"id": con.Id,
		}

		for key, valList := range c.tree {
			for keyVal, list := range valList {
				_, isOk := list.ConnectionList[con.Id]
				if isOk {
					entry[key] = keyVal
				}
			}
		}

		entries = append(entries, entry)
	}

	return entries
}
