package vault

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"
)

func TestEditLocationNonexisting(t *testing.T) {
	v, err := New("testpass")
	if err != nil {
		t.Fatal(err)
	}

	err = v.Edit("testlocation", Credential{"testusername", "testpassword"})
	if err != ErrNoSuchCredential {
		t.Fatal("expected Edit on non-existant location to return ErrNoSuchCredential")
	}
}

func TestEditLocation(t *testing.T) {
	v, err := New("testpass")
	if err != nil {
		t.Fatal(err)
	}

	err = v.Add("testlocation", Credential{"testusername", "testpassword"})
	if err != nil {
		t.Fatal(err)
	}

	err = v.Edit("testlocation", Credential{"testusername2", "testpassword2"})
	if err != nil {
		t.Fatal(err)
	}

	cred, err := v.Get("testlocation")
	if err != nil {
		t.Fatal(err)
	}

	if cred.Username != "testusername2" || cred.Password != "testpassword2" {
		t.Fatal("vault.Edit did not change credential data")
	}
}

func TestGetInvalidKey(t *testing.T) {
	v, err := New("testpass")
	if err != nil {
		t.Fatal(err)
	}
	v.secret = [32]byte{}
	if _, err = v.Get("test"); err != ErrCouldNotDecrypt {
		t.Fatal("expected v.Get to return ErrCouldNotDecrypt with invalid secret")
	}
}

func TestAddInvalidKey(t *testing.T) {
	v, err := New("testpass")
	if err != nil {
		t.Fatal(err)
	}
	v.secret = [32]byte{}
	if err = v.Add("testlocation", Credential{"test", "test2"}); err != ErrCouldNotDecrypt {
		t.Fatal("expected v.Add to return ErrCouldNotDecrypt with invalid secret")
	}
}

func TestHeavyVault(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	size := 10000

	v, err := New("testpass")
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < size; i++ {
		err = v.Add(fmt.Sprintf("testlocation%v", i), Credential{"testuser", "testpassword"})
		if err != nil {
			t.Fatal(err)
		}
	}

	err = v.Save("testvault.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("testvault.db")

	vopen, err := Open("testvault.db", "testpass")
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < size; i++ {
		cred, err := vopen.Get(fmt.Sprintf("testlocation%v", i))
		if err != nil {
			t.Fatal(err)
		}
		if cred.Username != "testuser" || cred.Password != "testpassword" {
			t.Fatal("huge vault did not contain testuser or testvault")
		}
	}
}

func TestNonexistentVaultOpen(t *testing.T) {
	_, err := Open("doesntexist.jpg", "nopass")
	if !os.IsNotExist(err) {
		t.Fatal("Open did not return IsNotExist for non-existant filename")
	}
}

func TestGenerate(t *testing.T) {
	v, err := New("testpass")
	if err != nil {
		t.Fatal(err)
	}
	if err = v.Generate("testlocation", "testusername"); err != nil {
		t.Fatal(err)
	}
	cred, err := v.Get("testlocation")
	if err != nil {
		t.Fatal(err)
	}
	if cred.Username != "testusername" {
		t.Fatal("Generate did not set username")
	}
	if cred.Password == "" {
		t.Fatal("generate did not generate a password")
	}
}

func TestGenerateExisting(t *testing.T) {
	v, err := New("testpass")
	if err != nil {
		t.Fatal(err)
	}
	err = v.Add("testlocation", Credential{"testuser", "testpass"})
	if err != nil {
		t.Fatal(err)
	}
	err = v.Generate("testlocation", "testuser")
	if err != ErrCredentialExists {
		t.Fatal("expected credential exists error on generate with existing location")
	}
}

func TestGetLocations(t *testing.T) {
	creds := []Credential{
		{"test1", "testpass1"},
		{"test2", "testpass2"},
		{"test3", "testpass3"},
	}
	locs := []string{"testloc1", "testloc2", "testloc3"}

	v, err := New("testpass")
	if err != nil {
		t.Fatal(err)
	}
	for i, cred := range creds {
		if err = v.Add(locs[i], cred); err != nil {
			t.Fatal(err)
		}
	}

	vaultLocations, err := v.Locations()
	if err != nil {
		t.Fatal(err)
	}

	sort.Strings(vaultLocations)
	if !reflect.DeepEqual(locs, vaultLocations) {
		t.Fatalf("expected %v to equal %v\n", vaultLocations, locs)
	}
}

func TestGetNonexisting(t *testing.T) {
	v, err := New("testpass")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = v.Get("testlocation"); err != ErrNoSuchCredential {
		t.Fatal("expected vault.Get on nonexisting credential to return ErrNoSuchCredential")
	}
}

func TestAddExisting(t *testing.T) {
	testCredential := Credential{"testuser", "testpass"}
	v, err := New("testpass")
	if err != nil {
		t.Fatal(err)
	}
	err = v.Add("testlocation", testCredential)
	if err != nil {
		t.Fatal(err)
	}
	err = v.Add("testlocation", testCredential)
	if err != ErrCredentialExists {
		t.Fatal("expected add on existing location to return ErrCredentialExists")
	}
}

func TestNewSaveOpen(t *testing.T) {
	testCredential := Credential{"testuser", "testpass"}

	v, err := New("testpass")
	if err != nil {
		t.Fatal(err)
	}
	err = v.Add("testlocation", testCredential)
	if err != nil {
		t.Fatal(err)
	}
	err = v.Save("pass.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("pass.db")

	vopen, err := Open("pass.db", "testpass")
	if err != nil {
		t.Fatal(err)
	}
	credential, err := vopen.Get("testlocation")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(&testCredential, credential) {
		t.Fatalf("vault did not store credential correctly. wanted %v got %v", testCredential, credential)
	}

	_, err = Open("pass.db", "wrongpass")
	if err != ErrCouldNotDecrypt {
		t.Fatal("Open decrypted given an incorrect passphrase")
	}
}

func TestNonceRotation(t *testing.T) {
	testCredential := Credential{"testuser", "testpass"}

	v, err := New("testpass")
	if err != nil {
		t.Fatal(err)
	}

	oldnonce := v.nonce
	oldsecret := v.secret

	v.Add("testlocation", testCredential)
	err = v.Save("pass.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("pass.db")

	vopen, err := Open("pass.db", "testpass")
	if err != nil {
		t.Fatal(err)
	}
	if vopen.secret == oldsecret {
		t.Fatal("opened vault had the same secret as the previous vault")
	}
	if vopen.nonce == oldnonce {
		t.Fatal("opened vault had the same nonce as the previous vault")
	}
}

func BenchmarkVaultAdd(b *testing.B) {
	v, err := New("testpass")
	if err != nil {
	}
	for i := 0; i < b.N; i++ {
		err = v.Add(fmt.Sprintf("testlocation%v", i), Credential{"testuser", "testpass"})
		if err != nil {
		}
	}
}
