package services

import (
	"bytes"
	"testing"

	"catalina-support/backend/modules/settings/repositories"
)

func TestAICredentialIsEncryptedAndAuthenticated(t *testing.T) {
	service := &Service{}
	service.SetAICredentialKey(bytes.Repeat([]byte{7}, 32))

	nonce, ciphertext, err := service.encryptAICredential("sk-secret")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(ciphertext, []byte("sk-secret")) {
		t.Fatal("the stored credential contains the plaintext")
	}
	row := repositories.AISettings{CredentialNonce: nonce, CredentialCiphertext: ciphertext}
	plain, err := service.decryptAICredential(row)
	if err != nil || plain != "sk-secret" {
		t.Fatalf("round trip = %q, %v", plain, err)
	}

	row.CredentialCiphertext[0] ^= 1
	if _, err := service.decryptAICredential(row); err == nil {
		t.Fatal("a modified credential was accepted")
	}
}
