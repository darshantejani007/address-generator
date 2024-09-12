package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joho/godotenv"
)

const filePath = "wallets.txt"

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	acceptable_pattern := os.Getenv("ACCEPTABLE_PATTERN") // Include 0x prefix in the regexp pattern
	final_pattern := os.Getenv("FINAL_PATTERN")           // Include 0x prefix in the regexp pattern
	if acceptable_pattern == "" || final_pattern == "" {
		fmt.Println("ACCEPTABLE_PATTERN & FINAL_PATTERN (regex) env variables required in the .env file")
		return
	}

	resetFile(filePath)
	runLoopInParallel(ctx, acceptable_pattern, final_pattern)
}

func runLoopInParallel(ctx context.Context, acceptable_pattern string, final_pattern string) {
	fmt.Println("Generating addresses with pattern : ", acceptable_pattern)
	acceptable_re := regexp.MustCompile(acceptable_pattern)
	start := time.Now()
	var wg sync.WaitGroup
	concurrencyLimit := getCores()
	semaphore := make(chan struct{}, concurrencyLimit)
	achievedFinal := false
	var counterLock sync.Mutex
	var counter int
	for i := 0; i < concurrencyLimit; i++ {

		wg.Add(1)
		semaphore <- struct{}{}

		go func(ctx context.Context) {
			defer wg.Done()

			for {
				if achievedFinal {
					break
				}
				if ctx.Err() != nil {
					break
				}
				counterLock.Lock()
				counter++
				counterLock.Unlock()

				privateKey, address := generateRandomWallet()

				if acceptable_re.MatchString(address) {
					insertLineAtFile(filePath, privateKey, address)
					final_re := regexp.MustCompile(final_pattern)
					if final_re.MatchString(address) {
						fmt.Println("Final Address: ", address)
						fmt.Println("Private Key: ", privateKey)
						achievedFinal = true
					}
				}
			}

			<-semaphore
		}(ctx)
	}

	wg.Wait()
	end := time.Now()
	seconds := end.Sub(start).Seconds()
	if seconds > 0 {
		fmt.Println("\nTotal addresses generated at Address per seconds rate : ", float64(counter)/seconds)
		fmt.Println("Total time taken: ", seconds)
		fmt.Println("Total addresses generated: ", counter)
	}
}

func generateRandomWallet() (string, string) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf("Failed to generate key: %v", err)
	}

	// Extract the private key in hexadecimal format
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := fmt.Sprintf("%x", privateKeyBytes)

	// Derive the public key from the private key
	publicKey := privateKey.Public().(*ecdsa.PublicKey)

	// Generate the Ethereum address
	address := crypto.PubkeyToAddress(*publicKey).Hex()

	return privateKeyHex, address
}

func insertLineAtFile(filePath string, privateKey string, address string) {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer f.Close()

	if _, err = f.WriteString(privateKey + "," + address + "\n"); err != nil {
		log.Fatalf("Failed to write to file: %v", err)
	}
}

func resetFile(filePath string) {
	if _, err := os.Stat(filePath); err == nil {
		err := os.Remove(filePath)
		if err != nil {
			log.Fatalf("Failed to remove file: %v", err)
		}
	}
}

func getCores() int {
	cores := os.Getenv("PARALLEL_CORES")
	coresInt, _ := strconv.Atoi(cores)
	if coresInt == 0 {
		coresInt = 8
	}
	return coresInt
}
