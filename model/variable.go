package model

type QUEUE_PRODUCT string

const (
	PRODUCT_TO_ELASTIC        QUEUE_PRODUCT = "product_to_elastic"
	UPDATE_PRODUCT_TO_ELASTIC QUEUE_PRODUCT = "update_product_to_elastic"
	DELETE_PRODUCT_TO_ELASTIC QUEUE_PRODUCT = "delete_product_to_elastic"

	UP_COUNT_WAREHOUSE QUEUE_PRODUCT = "up_count_warehouse"
)
