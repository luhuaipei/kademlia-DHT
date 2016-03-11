package kademlia

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
    	"time"
	mathrand "math/rand"
	"sss"
	"log"
)

type VanashingDataObject struct {
	AccessKey  int64
	Ciphertext []byte
	NumberKeys byte
	Threshold  byte
}

func GenerateRandomCryptoKey() (ret []byte) {
	for i := 0; i < 32; i++ {
		ret = append(ret, uint8(mathrand.Intn(256)))
	}
	return
}

func GenerateRandomAccessKey() (accessKey int64) {
    r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
    accessKey = r.Int63()
    return
}

func CalculateSharedKeyLocations(accessKey int64, count int64) (ids []ID) {
	r := mathrand.New(mathrand.NewSource(accessKey))
	ids = make([]ID, count)
	for i := int64(0); i < count; i++ {
		for j := 0; j < IDBytes; j++ {
			ids[i][j] = uint8(r.Intn(256))
		}
	}
	return
}

func encrypt(key []byte, text []byte) (ciphertext []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	ciphertext = make([]byte, aes.BlockSize+len(text))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], text)
	return
}

func decrypt(key []byte, ciphertext []byte) (text []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext is not long enough")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return ciphertext
}

func VanishData(kadem Kademlia, data []byte, numberKeys byte, threshold byte) (vdo VanashingDataObject) {
	//create a random cryptographic key(K) use  
	key := GenerateRandomCryptoKey()
	//encrypt data
	ciphertext := encrypt(key, data)
	//break the key, K, into multiple shares using sss.Split in sss/sss.go
		//sss.Split returns a map (key, value storage)
	keys_map, err := sss.Split(numberKeys, threshold, key)
	if err != nil{
		log.Fatal(err)
	}
	//Create an access key(L). 
	L := GenerateRandomAccessKey()
	//Store these keys in the DHT
	locations := CalculateSharedKeyLocations(L, int64(numberKeys))
	for i := range keys_map{
		all := append([]byte{i}, keys_map[i]...)
		kadem.DoIterativeStore(locations[i], all) 
	}
	res := VanashingDataObject{L, ciphertext, numberKeys, threshold}
	return res
}

func UnvanishData(kadem *Kademlia, vdo VanashingDataObject) (data []byte) {
	location := CalculateSharedKeyLocations(vdo.AccessKey, int64(vdo.NumberKeys))
	counter := 0
	var m map[byte] []byte
	m = make(map[byte][]byte)
	for i := range location{
		res, _ := kadem.DoIterativeFindValue(location[i])
		if res != "ERR"{
			counter += 1
			all := []byte(res)
			k := all[0]
			v := all[1:]
			m[k] = v
			if counter >= int(vdo.Threshold){
				break
			}
		}
	}
	key := sss.Combine(m)
	secret := decrypt(key, vdo.Ciphertext)
	return secret
}
