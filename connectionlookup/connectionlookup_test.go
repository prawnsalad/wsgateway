package connectionlookup

import "testing"

func TestAddConnection(t *testing.T) {
	lib, _ := NewConnectionLookup("")
	c := NewConnection("1", nil)
	lib.AddConnection(c, map[string]string{})

	if len(lib.connections) != 1 {
		t.Error("expected 1 connection")
	}
}

func TestAddConnectionWithTags(t *testing.T) {
	lib, _ := NewConnectionLookup("")
	c := NewConnection("1", nil)
	lib.AddConnection(c, map[string]string{
		"test": "value",
	})

	list, isOk := lib.tree["test"]
	if !isOk {
		t.Error("expected tag 'test' to exist")
	}

	_, isOk = list["value"]
	if !isOk {
		t.Error("expected tag value 'value' to exist")
	}
}

func TestRemoveConnection(t *testing.T) {
	lib, _ := NewConnectionLookup("")
	c := NewConnection("1", nil)
	lib.AddConnection(c, map[string]string{
		"test": "value",
	})

	if len(lib.connections) != 1 {
		t.Error("expected 1 connection")
	}

	lib.RemoveConnection(c)

	if len(lib.connections) != 0 {
		t.Error("expected 0 connections")
	}

	list, isOk := lib.tree["test"]
	if isOk {
		t.Error("expected tag 'test' to have been removed")
	}

	_, isOk = list["value"]
	if isOk {
		t.Error("expected tag value 'value' to have been removed")
	}
}

func TestAddMultipleConnections(t *testing.T) {
	lib, _ := NewConnectionLookup("")
	c := NewConnection("1", nil)
	lib.AddConnection(c, map[string]string{})

	c2 := NewConnection("2", nil)
	lib.AddConnection(c2, map[string]string{})

	if len(lib.connections) != 2 {
		t.Error("expected 2 connections")
	}
}


func TestRemoveConnectionWithMultipleConnections(t *testing.T) {
	lib, _ := NewConnectionLookup("")
	c := NewConnection("1", nil)
	lib.AddConnection(c, map[string]string{
		"test": "value",
		"single": "value",
	})

	c2 := NewConnection("2", nil)
	lib.AddConnection(c2, map[string]string{
		"test": "value",
	})


	if len(lib.connections) != 2 {
		t.Error("expected 2 connections")
	}

	lib.RemoveConnection(c)

	if len(lib.connections) != 1 {
		t.Error("expected 1 connection")
	}

	// The common tags should still exist
	list, isOk := lib.tree["test"]
	if !isOk {
		t.Error("expected tag 'test' to still exist")
	}

	_, isOk = list["value"]
	if !isOk {
		t.Error("expected tag value 'value' to still exist")
	}

	// The tag on the removed connection should have been removed
	_, isOk = lib.tree["single"]
	if isOk {
		t.Error("expected tag 'single' to have been removed")
	}
}

func TestGetConnectionWithTags(t *testing.T) {
	lib, _ := NewConnectionLookup("")
	c := NewConnection("1", nil)
	lib.AddConnection(c, map[string]string{
		"grouppid": "grp1",
		"userid": "1",
	})

	c2 := NewConnection("2", nil)
	lib.AddConnection(c2, map[string]string{
		"grouppid": "grp1",
		"userid": "2",
	})

	if len(lib.connections) != 2 {
		t.Error("expected 2 connections")
	}

	cons := lib.GetConnectionsWithKeys(map[string]string{
		"userid": "1",
	})

	if len(cons) != 1 {
		t.Error("expected to find 1 connection")
	}

	if cons[0].Id != "1" {
		t.Error("expected to find the correct connection with tag user1")
	}

	cons = lib.GetConnectionsWithKeys(map[string]string{
		"grouppid": "grp1",
	})

	if len(cons) != 2 {
		t.Error("expected to find 2 connections")
	}

	cons = lib.GetConnectionsWithKeys(map[string]string{
		"unknowntag": "unknown",
	})

	if len(cons) != 0 {
		t.Error("expected to find 0 connections")
	}
}
