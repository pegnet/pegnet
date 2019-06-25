package cmd

import (
	"fmt"

	"github.com/FactomProject/factom"
	"github.com/spf13/cobra"
)

// Custom arg validation methods
// TODO: Add context? https://golang.org/pkg/context/

// CombineCobraArgs allows the combination of multiple PositionalArgs
func CombineCobraArgs(funcs ...cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		for _, f := range funcs {
			err := f(cmd, args)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// CustomArgOrderValidationBuilder return an arg validator. The arg validator
// will validate cli arguments based on the validation functions provided
// in the order of the validation functions.
//		Params:
//			strict		Enforce the number of args == number of validation funcs
//			valids		Validation functions
func CustomArgOrderValidationBuilder(strict bool, valids ...func(cmd *cobra.Command, args string) error) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if strict && len(valids) != len(args) {
			return fmt.Errorf("accepts %d arg(s), received %d", len(valids), len(args))
		}

		for i, arg := range args {
			if err := valids[i](cmd, arg); err != nil {
				return err
			}
		}
		return nil
	}
}

// ArgValidatorExample is just a template
//		Params:
//			cmd
//			arg		Arg to validate
func ArgValidatorExample(cmd *cobra.Command, arg string) error {
	return nil
}

// ArgValidatorExists checks that arg exists
func ArgValidatorExists(cmd *cobra.Command, arg string) error {
	return nil
}

// ArgValidatorFCTAddress checks for FCT address
func ArgValidatorFCTAddress(cmd *cobra.Command, arg string) error {
	if arg[:2] != "FA" {
		return fmt.Errorf("FCT addresses start with FA")
	}
	if factom.IsValidAddress(arg) {
		return nil
	}
	return fmt.Errorf("%s is not a valid FCT address", arg)
}
