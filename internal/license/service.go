package license

type Service struct {
	crypto *Crypto
}

func NewService(crypto *Crypto) *Service {
	return &Service{crypto: crypto}
}

func (s *Service) DecryptBundle(fileContent string) (Bundle, error) {
	return s.crypto.DecryptBundle(fileContent)
}

func (s *Service) LicenseDetail(license DecodedLicense) map[string]any {
	verified := make([]map[string]any, len(license.Tokens))
	for i, token := range license.Tokens {
		item := map[string]any{"index": i + 1, "isVerified": false, "data": nil, "error": nil}
		data, err := s.crypto.VerifyToken(token)
		if err != nil {
			item["error"] = err.Error()
		} else if !license.IsActive {
			item["error"] = "LICENSE_INACTIVE"
		} else {
			item["isVerified"] = true
			item["data"] = data
		}
		verified[i] = item
	}
	meta := map[string]any{
		"company": license.Company, "licenseType": license.LicenseType, "expiry": license.Expiry,
		"issuedAt": license.IssuedAt, "importedAt": license.CreatedAt,
	}
	if license.ID != "" {
		meta["isActive"] = license.IsActive
	}
	return map[string]any{"meta": meta, "tokens": verified}
}
