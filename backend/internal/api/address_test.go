package api

import "testing"

func TestValidAddressFields(t *testing.T) {
	valid := addressRequest{
		Recipient: "张女士",
		Phone:     "+86 13800138000",
		Province:  "贵州省",
		City:      "贵阳市",
		District:  "南明区",
		Detail:    "花果园一号楼",
	}
	if !validAddressFields(valid) {
		t.Fatal("expected complete address to be valid")
	}

	invalid := valid
	invalid.Phone = "      "
	if validAddressFields(invalid) {
		t.Fatal("expected phone without digits to be invalid")
	}

	invalid = valid
	invalid.District = ""
	if validAddressFields(invalid) {
		t.Fatal("expected address without a district to be invalid")
	}
}
