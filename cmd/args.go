// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"encoding/hex"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/polling"
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

func ArgValidatorFCTAmount(cmd *cobra.Command, arg string) error {
	// The FCT amount must not be beyond 1e8 divisible
	_, err := factoidToFactoshi(arg)
	return err
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
	list := append([]string{"all"}, common.AllAssets...)
	for _, an := range list {
		if strings.ToLower(arg) == strings.ToLower(an) {
			return nil
		}
	}
	return fmt.Errorf("not a valid asset. Options include: %v", list)
}

// ArgValidatorAssetOrExchange checks for an asset or datasource name
func ArgValidatorAssetOrExchange(cmd *cobra.Command, arg string) error {
	list := append(common.AllAssets, polling.AllDataSourcesList()...)
	for _, an := range list {
		if strings.ToLower(arg) == strings.ToLower(an) {
			return nil
		}
	}

	return fmt.Errorf("not a valid asset/datasource. Options include: %v", list)
}

// ArgValidatorAsset checks for valid asset
func ArgValidatorAsset(cmd *cobra.Command, arg string) error {
	for _, an := range common.AllAssets {
		if strings.ToLower(arg) == strings.ToLower(an) {
			return nil
		}
	}
	return fmt.Errorf("not a valid asset. Assets include: %v", common.AllAssets)
}

// ArgValidatorHexHash checks for valid hash
func ArgValidatorHexHash(cmd *cobra.Command, arg string) error {
	if len(arg) != 64 {
		return fmt.Errorf("a hash is a 64 character hex string")
	}
	_, err := hex.DecodeString(arg)
	return err
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

// FactoidToFactoshi is taken from the factom lib, but errors when extra decimals provided
func factoidToFactoshi(amt string) (uint64, error) {
	valid := regexp.MustCompile(`^([0-9]+)?(\.[0-9]+)?$`)
	if !valid.MatchString(amt) {
		return 0, nil
	}

	var total uint64 = 0

	dot := regexp.MustCompile(`\.`)
	pieces := dot.Split(amt, 2)
	whole, _ := strconv.Atoi(pieces[0])
	total += uint64(whole) * 1e8

	if len(pieces) > 1 {
		if len(pieces[1]) > 8 {
			return 0, fmt.Errorf("factoids are only subdivisible up to 1e-8, trim back on the number of decimal places")
		}

		a := regexp.MustCompile(`(0*)([0-9]+)$`)

		as := a.FindStringSubmatch(pieces[1])
		part, _ := strconv.Atoi(as[0])
		power := len(as[1]) + len(as[2])
		total += uint64(part * 1e8 / int(math.Pow10(power)))
	}

	return total, nil
}
