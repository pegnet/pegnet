package common

import (
	"bytes"
	"encoding/hex"
	"reflect"
	"testing"
)

var validRCD = map[string][]byte{
	"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu": []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	"FA3upjWMKHmStAHR5ZgKVK4zVHPb8U74L2wzKaaSDQEonHajiLeq": []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
}
var testAddresses = map[string]map[string]string{
	"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu": map[string]string{
		"pPNT":  "pPNT_111111111111111111111111111111112G1Pfj",
		"tPNT":  "tPNT_111111111111111111111111111111112fZAPP",
		"pUSD":  "pUSD_111111111111111111111111111111112gryZ2",
		"tUSD":  "tUSD_111111111111111111111111111111114x5nT3",
		"pEUR":  "pEUR_111111111111111111111111111111116wykv8",
		"tEUR":  "tEUR_111111111111111111111111111111116tWRvU",
		"pJPY":  "pJPY_111111111111111111111111111111112sHenP",
		"tJPY":  "tJPY_111111111111111111111111111111115PDu2w",
		"pGBP":  "pGBP_111111111111111111111111111111115BM67",
		"tGBP":  "tGBP_111111111111111111111111111111114sS44V",
		"pCAD":  "pCAD_111111111111111111111111111111115cJJ1q",
		"tCAD":  "tCAD_111111111111111111111111111111112WwWg2",
		"pCHF":  "pCHF_111111111111111111111111111111115Gkuri",
		"tCHF":  "tCHF_1111111111111111111111111111111167yNQj",
		"pINR":  "pINR_111111111111111111111111111111113sGbg6",
		"tINR":  "tINR_111111111111111111111111111111115tjct9",
		"pSGD":  "pSGD_1111111111111111111111111111111176x5uD",
		"tSGD":  "tSGD_111111111111111111111111111111116T512L",
		"pCNY":  "pCNY_111111111111111111111111111111114iQTuz",
		"tCNY":  "tCNY_111111111111111111111111111111114xirFd",
		"pHKD":  "pHKD_111111111111111111111111111111114gm2HY",
		"tHKD":  "tHKD_111111111111111111111111111111113mtpjp",
		"pKRW":  "pKRW_111111111111111111111111111111115WXEDy",
		"tKRW":  "tKRW_111111111111111111111111111111115kbS3F",
		"pBRL":  "pBRL_111111111111111111111111111111112jn2xG",
		"tBRL":  "tBRL_111111111111111111111111111111112YSsfY",
		"pPHP":  "pPHP_111111111111111111111111111111116R7mDD",
		"tPHP":  "tPHP_111111111111111111111111111111115VoirA",
		"pMXN":  "pMXN_111111111111111111111111111111114jNFH8",
		"tMXN":  "tMXN_111111111111111111111111111111114sCUpR",
		"pXAU":  "pXAU_111111111111111111111111111111115XNTbU",
		"tXAU":  "tXAU_111111111111111111111111111111116L6JHV",
		"pXAG":  "pXAG_111111111111111111111111111111114MaSev",
		"tXAG":  "tXAG_1111111111111111111111111111111141n3te",
		"pXPD":  "pXPD_111111111111111111111111111111114ywbbz",
		"tXPD":  "tXPD_111111111111111111111111111111113ss5bM",
		"pXPT":  "pXPT_111111111111111111111111111111116eH2cG",
		"tXPT":  "tXPT_111111111111111111111111111111115Af8K3",
		"pXBT":  "pXBT_111111111111111111111111111111113jGc2w",
		"tXBT":  "tXBT_111111111111111111111111111111116e4MA1",
		"pETH":  "pETH_111111111111111111111111111111116YGV1u",
		"tETH":  "tETH_111111111111111111111111111111114NgcZE",
		"pLTC":  "pLTC_111111111111111111111111111111117ArWiq",
		"tLTC":  "tLTC_1111111111111111111111111111111158gYaC",
		"pRVN":  "pRVN_111111111111111111111111111111116thTWe",
		"tRVN":  "tRVN_111111111111111111111111111111114eirns",
		"pXBC":  "pXBC_111111111111111111111111111111117Mwz34",
		"tXBC":  "tXBC_111111111111111111111111111111117UGSSg",
		"pFCT":  "pFCT_111111111111111111111111111111115oz8WJ",
		"tFCT":  "tFCT_111111111111111111111111111111112Vc4kC",
		"pBNB":  "pBNB_11111111111111111111111111111111zTiAw",
		"tBNB":  "tBNB_111111111111111111111111111111112GezAh",
		"pXLM":  "pXLM_111111111111111111111111111111113WE82d",
		"tXLM":  "tXLM_111111111111111111111111111111115SpuRv",
		"pADA":  "pADA_111111111111111111111111111111114VhJDk",
		"tADA":  "tADA_111111111111111111111111111111116jVTzd",
		"pXMR":  "pXMR_11111111111111111111111111111111Puiof",
		"tXMR":  "tXMR_11111111111111111111111111111111uEWPN",
		"pDASH": "pDASH_1111111111111111111111111111111132A489",
		"tDASH": "tDASH_111111111111111111111111111111116AA5zX",
		"pZEC":  "pZEC_11111111111111111111111111111111tBGDN",
		"tZEC":  "tZEC_111111111111111111111111111111112JThxB",
		"pDCR":  "pDCR_111111111111111111111111111111117E7UjW",
		"tDCR":  "tDCR_111111111111111111111111111111115Qoowu",
	},
	"FA3upjWMKHmStAHR5ZgKVK4zVHPb8U74L2wzKaaSDQEonHajiLeq": map[string]string{
		"pPNT":  "pPNT_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVaSSkVB",
		"tPNT":  "tPNT_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVaF5AbG",
		"pUSD":  "pUSD_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVay8ADs",
		"tUSD":  "tUSD_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVZ8PMSu",
		"pEUR":  "pEUR_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVWfnJtY",
		"tEUR":  "tEUR_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVcPhjwy",
		"pJPY":  "pJPY_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVcxgXM5",
		"tJPY":  "tJPY_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVXUA6qC",
		"pGBP":  "pGBP_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVbRT3Lp",
		"tGBP":  "tGBP_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVZDNCvb",
		"pCAD":  "pCAD_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVcgqyxn",
		"tCAD":  "tCAD_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVchA71L",
		"pCHF":  "pCHF_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVYAzjDf",
		"tCHF":  "tCHF_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVbNA8TF",
		"pINR":  "pINR_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVapc7Hm",
		"tINR":  "tINR_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVXtcuQe",
		"pSGD":  "pSGD_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVZYkTFa",
		"tSGD":  "tSGD_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVXA1BnX",
		"pCNY":  "pCNY_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVcV5yNH",
		"tCNY":  "tCNY_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVWnrcRn",
		"pHKD":  "pHKD_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVYWpKxT",
		"tHKD":  "tHKD_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVcJfKQW",
		"pKRW":  "pKRW_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVbUcL1Z",
		"tKRW":  "tKRW_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVcRTNph",
		"pBRL":  "pBRL_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVcYDjJt",
		"tBRL":  "tBRL_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVWcJt46",
		"pPHP":  "pPHP_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVb391jr",
		"tPHP":  "tPHP_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVYNocUe",
		"pMXN":  "pMXN_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVaY1Qk8",
		"tMXN":  "tMXN_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVctv6Bn",
		"pXAU":  "pXAU_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVX5AkND",
		"tXAU":  "tXAU_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVYEroX3",
		"pXAG":  "pXAG_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVay9gLB",
		"tXAG":  "tXAG_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVXEjtRA",
		"pXPD":  "pXPD_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVcP3j7A",
		"tXPD":  "tXPD_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVctupu5",
		"pXPT":  "pXPT_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVbsrk4P",
		"tXPT":  "tXPT_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVbqykdc",
		"pXBT":  "pXBT_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVcaDe4W",
		"tXBT":  "tXBT_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVaz6Fzb",
		"pETH":  "pETH_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVcac2Ny",
		"tETH":  "tETH_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVbTt3iY",
		"pLTC":  "pLTC_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVbjRLPE",
		"tLTC":  "tLTC_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVXvTQtC",
		"pRVN":  "pRVN_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVbtW6y4",
		"tRVN":  "tRVN_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVZ9X1Et",
		"pXBC":  "pXBC_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVcXApzv",
		"tXBC":  "tXBC_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVY3DeBu",
		"pFCT":  "pFCT_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVbw9kcp",
		"tFCT":  "tFCT_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVcrQvbW",
		"pBNB":  "pBNB_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVa13RWy",
		"tBNB":  "tBNB_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVZ4YNC6",
		"pXLM":  "pXLM_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVYNYmFZ",
		"tXLM":  "tXLM_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVX5pkUs",
		"pADA":  "pADA_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVamapco",
		"tADA":  "tADA_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVb31vz3",
		"pXMR":  "pXMR_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVa6Pm5L",
		"tXMR":  "tXMR_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVYcLUZG",
		"pDASH": "pDASH_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVcu9hQV",
		"tDASH": "tDASH_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVZbPjdX",
		"pZEC":  "pZEC_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVYod8pM",
		"tZEC":  "tZEC_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVY941ex",
		"pDCR":  "pDCR_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVXinSmp",
		"tDCR":  "tDCR_2wkBET2rRgE8pahuaczxKbmv7ciehqsne57F9gtzf1PVZNBdnF",
	},
}

/* **** TEST ADDRESSES WERE GENERATED WITH THIS, USING THE OLD FUNCTION ****
func buildTestAddresses() {
	fa := []string{
		"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu",
		"FA3upjWMKHmStAHR5ZgKVK4zVHPb8U74L2wzKaaSDQEonHajiLeq",
	}

	fmt.Println("var validRCD = map[string][]byte{")
	for _, s := range fa {
		fmt.Printf("\"%s\": []byte{", s)
		raw := base58.Decode(s)[2:34]
		for i, b := range raw {
			if i != 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%d", b)
		}
		fmt.Println("},")
	}
	fmt.Println("}")
	fmt.Println("var testAddresses = map[string]map[string]string{")
	hex.EncodeToString(nil)

	for _, addr := range fa {
		assets, err := ConvertUserFctToUserPegNetAssets(addr)
		if err != nil {
			panic(err)
		}

		fmt.Printf("\t\"%s\": map[string]string{\n", addr)
		for _, s := range assets {
			split := strings.Split(s, "_")
			fmt.Printf("\t\t\"%s\": \"%s\",\n", split[0], s)
		}
		fmt.Println("\t},")
	}

	fmt.Println("}")
}*/

func TestAddress_All(t *testing.T) {
	for addr, assets := range testAddresses {
		fa, err := ParseAddress(addr)
		if err != nil {
			t.Errorf("unable to parse %s into factom address: %v", addr, err)
		}

		if orig := fa.String(); orig != addr {
			t.Errorf("fa.String() did not return the same value: want = %s, got = %s", addr, orig)
		}

		if bytes.Compare(fa.RCD, validRCD[addr]) != 0 {
			t.Errorf("fa %s did not decode the valid RCD: want = %s, got %s", addr, hex.EncodeToString(validRCD[addr]), hex.EncodeToString(fa.RCD))
		}

		for asset, paddr := range assets {
			a, err := ParseAddress(paddr)
			if err != nil {
				t.Errorf("error converting %s to address: %v", paddr, err)
			}

			if orig := a.String(); orig != paddr {
				t.Errorf("a.String() for %s returned %s", paddr, orig)
			}

			if a.Prefix != asset {
				t.Errorf("parsed the wrong asset for %s: want = %s, got = %s", paddr, asset, a.Prefix)
			}

			if !a.IsSameBase(fa) {
				t.Errorf("asset %s did not decode the valid RCD: want = %s, got %s", paddr, hex.EncodeToString(validRCD[addr]), hex.EncodeToString(a.RCD))
			}

			if fact := a.FactomAddress(); fact != addr {
				t.Errorf("%s did not convert to factom address: want = %s, got = %s", paddr, addr, fact)
			}

			for prefix, otheraddr := range assets { // includes self check
				if other := a.ToAsset(prefix); other != otheraddr {
					t.Errorf("converting %s to %s did not match: want = %s, got = %s", paddr, prefix, otheraddr, other)
				}
			}
		}
	}
}

func TestAddress_SetPrefix(t *testing.T) {
	a := new(Address)
	type args struct {
		prefix string
	}
	tests := []struct {
		name    string
		a       *Address
		args    args
		wantErr bool
	}{
		{"empty", a, args{""}, false},
		{"one letter", a, args{"a"}, true},
		{"network token", a, args{"pPNT"}, false},
		{"testnetwork token", a, args{"tPNT"}, false},
		{"wrong network", a, args{"dPNT"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.a.SetPrefix(tt.args.prefix); (err != nil) != tt.wantErr {
				t.Errorf("Address.SetPrefix() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseAddress(t *testing.T) {
	//test := "FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu"
	min := "pPNT_111111111111111111111111111111112G1Pfj"
	max := "pDASH_1111111111111111111111111111111132A489"

	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    *Address
		wantErr bool
	}{ // only testing errors, the valid ones are covered above
		{"empty", args{""}, nil, true},
		{"too short", args{min[:len(min)-1]}, nil, true},
		{"too long", args{max + "1"}, nil, true},
		{"invalid checksum fa", args{"FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgb"}, nil, true},
		{"invalid checksum asset", args{"pPNT_111111111111111111111111111111112G1Pfk"}, nil, true},
		{"invalid asset", args{"pFCT_111111111111111111111111111111112G1Pfj"}, nil, true},
		{"nonexisting asset", args{"pFOO_111111111111111111111111111111112G1Pfj"}, nil, true},
		{"invalid Base58", args{"FA0y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAddress(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
