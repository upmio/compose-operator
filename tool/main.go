package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/upmio/compose-operator/pkg/utils"
)

func main() {
	var (
		key       = flag.String("key", "", "AES encryption key (32 characters required)")
		plaintext = flag.String("plaintext", "", "Plaintext to encrypt")
		username  = flag.String("username", "", "Username for binary file output (used with -plaintext)")
		file      = flag.String("file", "", "Binary file to decrypt (used with -key only)")
		decrypt   = flag.String("decrypt", "", "Ciphertext to decrypt (base64 encoded)")
		help      = flag.Bool("help", false, "Show help information")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "AES-CTR Encryption Tool\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s -key <32-char-key> -plaintext <text> -username <name>  # Encrypt and save to <name>.bin\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -key <32-char-key> -file <filename>                    # Decrypt binary file\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -key <32-char-key> -decrypt <base64>                   # Decrypt base64 text\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Encrypt password and save to mysql.bin\n")
		fmt.Fprintf(os.Stderr, "  %s -key \"your-32-character-aes-key-here\" -plaintext \"mypassword\" -username \"mysql\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n  # Decrypt from binary file\n")
		fmt.Fprintf(os.Stderr, "  %s -key \"your-32-character-aes-key-here\" -file \"mysql.bin\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n  # Decrypt base64 ciphertext\n")
		fmt.Fprintf(os.Stderr, "  %s -key \"your-32-character-aes-key-here\" -decrypt \"base64-encoded-ciphertext\"\n", os.Args[0])
	}

	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	// Validate key
	if *key == "" {
		fmt.Fprintf(os.Stderr, "Error: AES key is required\n")
		flag.Usage()
		os.Exit(1)
	}

	if len(*key) != 32 {
		fmt.Fprintf(os.Stderr, "Error: AES key must be exactly 32 characters, got %d characters\n", len(*key))
		os.Exit(1)
	}

	// Set the AES key
	os.Setenv("AES_SECRET_KEY", *key)
	if err := utils.ValidateAndSetAESKey(); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting AES key: %v\n", err)
		os.Exit(1)
	}

	// Handle encryption
	if *plaintext != "" {
		if *decrypt != "" || *file != "" {
			fmt.Fprintf(os.Stderr, "Error: Cannot specify -plaintext with -decrypt or -file\n")
			os.Exit(1)
		}

		if *username == "" {
			fmt.Fprintf(os.Stderr, "Error: -username is required when using -plaintext\n")
			os.Exit(1)
		}

		encryptedBytes, err := utils.AES_CTR_EncryptToBytes([]byte(*plaintext))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Encryption failed: %v\n", err)
			os.Exit(1)
		}

		filename := *username + ".bin"
		err = ioutil.WriteFile(filename, encryptedBytes, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write file %s: %v\n", filename, err)
			os.Exit(1)
		}

		fmt.Printf("Plaintext: %s\n", *plaintext)
		fmt.Printf("Encrypted and saved to: %s\n", filename)
		return
	}

	// Handle file decryption
	if *file != "" {
		if *decrypt != "" {
			fmt.Fprintf(os.Stderr, "Error: Cannot specify both -file and -decrypt\n")
			os.Exit(1)
		}

		encryptedBytes, err := ioutil.ReadFile(*file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read file %s: %v\n", *file, err)
			os.Exit(1)
		}

		decrypted, err := utils.AES_CTR_DecryptFromBytes(encryptedBytes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Decryption failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("File: %s\n", *file)
		fmt.Printf("Decrypted: %s\n", string(decrypted))
		return
	}

	// Handle base64 decryption
	if *decrypt != "" {
		decrypted, err := utils.AES_CTR_Decrypt(*decrypt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Decryption failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Ciphertext: %s\n", *decrypt)
		fmt.Printf("Decrypted: %s\n", string(decrypted))
		return
	}

	// No action specified
	fmt.Fprintf(os.Stderr, "Error: Must specify -plaintext with -username, -file, or -decrypt\n")
	flag.Usage()
	os.Exit(1)
}