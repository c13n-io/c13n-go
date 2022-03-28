package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

type promptContent struct {
	errorMsg string
	label    string
}

func promptGetInput(pc promptContent) string {
	validate := func(input string) error {
		if len(input) <= 0 {
			return errors.New(pc.errorMsg)
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    pc.label,
		Validate: validate,
	}

	result, err := prompt.Run()
	if err != nil {
		logger.WithError(err).Error("Prompt failed")
		os.Exit(1)
	}

	return result
}

// generateHashCmd represents the generateHash command
var generateHashCmd = &cobra.Command{
	Use:   "generateHash",
	Short: "Generate bcrypt hash from provided password.",

	Run: func(cmd *cobra.Command, args []string) {
		cost, err := cmd.Flags().GetInt("cost")
		if err != nil {
			logger.WithError(err).Error()
			os.Exit(1)
		}

		if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
			logger.Errorf("Cost outside of range [%d,%d]", bcrypt.MinCost, bcrypt.MaxCost)
			os.Exit(1)
		}

		passwordInputPrompt := promptContent{
			"Please provide your password.",
			"Password",
		}
		passwordInput := promptGetInput(passwordInputPrompt)

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordInput), cost)
		if err != nil {
			logger.WithError(err).Error()
			os.Exit(1)
		}
		fmt.Println(string(hashedPassword))
	},
}

func init() {
	rootCmd.AddCommand(generateHashCmd)

	generateHashFlags := generateHashCmd.Flags()
	generateHashFlags.Int("cost", 10, "Default cost for bcrypt hashing")
	err := generateHashCmd.MarkFlagRequired("cost")
	if err != nil {
		logger.WithError(err).Error()
		os.Exit(1)
	}
}
