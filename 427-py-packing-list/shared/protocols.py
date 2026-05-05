from typing import Final

API_PREFIX: Final[str] = "/api/v1"

ORDERS_ENDPOINT: Final[str] = f"{API_PREFIX}/orders"
ORDER_ENDPOINT: Final[str] = f"{API_PREFIX}/orders/{{order_id}}"

PACKING_LISTS_ENDPOINT: Final[str] = f"{API_PREFIX}/orders/{{order_id}}/packing"

PACKAGES_ENDPOINT: Final[str] = f"{API_PREFIX}/orders/{{order_id}}/packages"
PACKAGE_ENDPOINT: Final[str] = f"{API_PREFIX}/packages/{{package_id}}"
PACKAGE_ITEMS_ENDPOINT: Final[str] = f"{API_PREFIX}/packages/{{package_id}}/items"
PACKAGE_TRACKING_ENDPOINT: Final[str] = f"{API_PREFIX}/packages/{{package_id}}/tracking"

PACKING_TEMPLATES_ENDPOINT: Final[str] = f"{API_PREFIX}/packing-templates"
PACKING_TEMPLATE_ENDPOINT: Final[str] = f"{API_PREFIX}/packing-templates/{{template_id}}"

BOX_TYPES_ENDPOINT: Final[str] = f"{API_PREFIX}/box-types"
BOX_TYPE_ENDPOINT: Final[str] = f"{API_PREFIX}/box-types/{{box_type_id}}"

PACKING_SUGGESTIONS_ENDPOINT: Final[str] = f"{API_PREFIX}/orders/{{order_id}}/packing-suggestions"

PACKING_LIST_FORMAT_ENDPOINT: Final[str] = f"{API_PREFIX}/orders/{{order_id}}/packing-list-text"

MATERIAL_STATS_ENDPOINT: Final[str] = f"{API_PREFIX}/material-stats"
