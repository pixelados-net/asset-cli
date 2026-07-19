package pets

import "testing"

func TestNewRealmCommandExposesCheck(t *testing.T) {
	realm := NewRealmCommand()
	if realm.Use != "pets" || len(realm.Commands()) != 1 || realm.Commands()[0].Name() != "check" {
		t.Fatalf("realm = %#v, commands = %#v", realm.Use, realm.Commands())
	}
}
