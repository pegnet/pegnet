package cmd

import (
	"fmt"
	"strings"

	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/common"
	"github.com/spf13/cobra"
)

//
// Custom arg validation methods
//
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
	if len(arg) > 2 && arg[:2] != "FA" {
		return fmt.Errorf("FCT addresses start with FA")
	}
	if factom.IsValidAddress(arg) {
		return nil
	}
	return fmt.Errorf("%s is not a valid FCT address", arg)
}

// ArgValidatorECAddress checks for EC address
func ArgValidatorECAddress(cmd *cobra.Command, arg string) error {
	if len(arg) > 2 && arg[:2] != "EC" {
		return fmt.Errorf("EC addresses start with EC")
	}
	if factom.IsValidAddress(arg) {
		return nil
	}
	return fmt.Errorf("%s is not a valid EC address", arg)
}

// ArgValidatorAssetAndAll checks for valid asset or 'all'
func ArgValidatorAssetAndAll(cmd *cobra.Command, arg string) error {
	list := append([]string{"all"}, common.AssetNames...)
	for _, an := range list {
		if strings.ToLower(arg) == strings.ToLower(an) {
			return nil
		}
	}
	return fmt.Errorf("Not a valid asset. Options include: %v", list)
}

// ArgValidatorAsset checks for valid asset
func ArgValidatorAsset(cmd *cobra.Command, arg string) error {
	for _, an := range common.AssetNames {
		if strings.ToLower(arg) == strings.ToLower(an) {
			return nil
		}
	}
	return fmt.Errorf("Not a valid asset. Assets include: %v", common.AssetNames)
}

// Custom Completion args

func ValidOwnedFCTAddresses() []string {
	fas, _, err := factom.FetchAddresses()
	if err != nil {
		return []string{""}
	}
	var strs []string
	for _, fa := range fas {
		strs = append(strs, fa.String())
	}
	return strs
}

func ValidOwnedECAddresses() []string {
	_, ecs, err := factom.FetchAddresses()
	if err != nil {
		return []string{""}
	}
	var strs []string
	for _, ec := range ecs {
		strs = append(strs, ec.String())
	}
	return strs
}
