package bitcoin

import (
	"testing"
)

func TestCreateTransaction(t *testing.T) {
	expected := "4e8378675bcf6a389c8cfe246094aafa44249e48ab88a40e6fda3bf0f44f916a"

	res, err := CreateTransactionNew("5HusYj2b2x4nroApgfvaSfKYZhRbKFH41bVyPooymbC6KfgSXdD", []Destination{
		{
			Addr:   "1KKKK6N21XKo48zWKuQKXdvSsCf95ibHFa",
			Amount: int64(91234),
		},
	},
		[]string{
			"81b4c832d70cb56ff957589752eb4125a4cab78a25a8fc52d6a09e5bd4404d48",
		}, false)
	if err != nil {
		t.Logf("crate failed, error: '%v'\n", err)
		t.FailNow()
	}

	if res.TxId != expected {
		t.Logf("new result hash: '%s'\n", res.TxId)
		t.FailNow()
	}

	t.Logf("success\n")
}
