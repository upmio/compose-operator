package cmd

import (
	"errors"
	"fmt"
	"github.com/upmio/compose-operator/pkg/utils"
	"os"

	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "tool for encrypt and decrypt by AES-256-CTR",
	RunE: func(cmd *cobra.Command, args []string) error {

		return errors.New("No flags find")
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		//fmt.Println(err)
		os.Exit(1)
	}
}

var (
	plaintext, key, file, username string
)

func init() {
	rootCmd.AddCommand(decryptCmd)
	rootCmd.AddCommand(encryptCmd)

	encryptCmd.PersistentFlags().StringVarP(&plaintext, "plaintext", "p", "", "Plaintext to encrypt")
	encryptCmd.PersistentFlags().StringVarP(&key, "key", "k", "", "AES encryption key (32 characters required)")
	encryptCmd.PersistentFlags().StringVarP(&username, "username", "u", "", "Username for binary file output")

	decryptCmd.PersistentFlags().StringVarP(&file, "file", "f", "", "Binary file to decrypt")
	decryptCmd.PersistentFlags().StringVarP(&key, "key", "k", "", "AES encryption key (32 characters required)")
}

// RootCmd represents the base command when called without any subcommands
var decryptCmd = &cobra.Command{
	Use: "decrypt",
	RunE: func(cmd *cobra.Command, args []string) error {

		// Validate key
		if key == "" {
			return fmt.Errorf("AES key is required")
		}

		if len(key) != 32 {
			return fmt.Errorf("AES key must be exactly 32 characters, got %d characters", len(key))
		}

		// Set the AES key
		err := os.Setenv(utils.AESKeyEnvVar, key)
		if err != nil {
			return err
		}

		if err := utils.ValidateAndSetAESKey(); err != nil {
			return err
		}

		encryptedBytes, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("Failed to read file %s: %v\n", file, err)
		}

		decrypted, err := utils.AES_CTR_Decrypt(encryptedBytes)
		if err != nil {
			return fmt.Errorf("Failed to decrypt file %s: %v\n", file, err)
		}

		fmt.Printf("File: %s\n", file)
		fmt.Printf("Decrypted: %s\n", string(decrypted))

		return nil
	},
}

// RootCmd represents the base command when called without any subcommands
var encryptCmd = &cobra.Command{
	Use: "encrypt",
	RunE: func(cmd *cobra.Command, args []string) error {

		// Validate key
		if key == "" {
			return fmt.Errorf("AES key is required")
		}

		if len(key) != 32 {
			return fmt.Errorf("AES key must be exactly 32 characters, got %d characters", len(key))
		}

		// Set the AES key
		err := os.Setenv(utils.AESKeyEnvVar, key)
		if err != nil {
			return err
		}

		if err := utils.ValidateAndSetAESKey(); err != nil {
			return err
		}

		encryptedBytes, err := utils.AES_CTR_Encrypt([]byte(plaintext))
		if err != nil {
			return err
		}

		filename := fmt.Sprintf("%s.bin", username)
		err = os.WriteFile(filename, encryptedBytes, 0644)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %v", filename, err)
		}

		fmt.Printf("Plaintext: %s\n", plaintext)
		fmt.Printf("Encrypted and saved to: %s\n", filename)

		return nil
	},
}
