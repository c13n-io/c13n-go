package cmd

import (
	"errors"
	"fmt"

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
		logger.WithError(err).Fatal("Prompt failed")
	}

	return result
}

// generateHashCmd represents the generateHash command
var genpwdhash = &cobra.Command{
	Use:   "genpwdhash",
	Short: "Generate bcrypt hash from provided password for RPC authentication",
	Long: "Generate bcrypt hash from provided password. " +
		"This can be used as a value for the server.pwdhash config " +
		"field or --server-pwdhash CLI option",

	Run: func(cmd *cobra.Command, args []string) {
		cost, err := cmd.Flags().GetInt("cost")
		if err != nil {
			logger.WithError(err).Fatal("could not get cost value")
		}

		if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
			logger.Fatalf("cost outside of range [%d,%d]", bcrypt.MinCost, bcrypt.MaxCost)
		}

		passwordInputPrompt := promptContent{
			"Please provide your password.",
			"Password",
		}
		passwordInput := promptGetInput(passwordInputPrompt)

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordInput), cost)
		if err != nil {
			logger.WithError(err).Fatal("could not generate bcrypt hash")
		}
		fmt.Println(string(hashedPassword))
	},
}

func init() {
	rootCmd.AddCommand(genpwdhash)

	generateHashFlags := genpwdhash.Flags()
	generateHashFlags.Int("cost", 10, "Default cost for bcrypt hashing")
}
