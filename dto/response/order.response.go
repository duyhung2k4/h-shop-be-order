package response

import "app/model"

type OrderResponse struct {
	GroupOrder model.GroupOrder `json:"groupOrder"`
	VnpHref    string           `json:"vnpHref"`
}
