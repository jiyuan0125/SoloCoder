from typing import Any, Optional, cast
import httpx

from shared import protocols


DEFAULT_BASE_URL = "http://localhost:8000"


class APIClient:
    def __init__(self, base_url: Optional[str] = None) -> None:
        self.base_url = base_url or DEFAULT_BASE_URL

    def _get_client(self) -> httpx.Client:
        return httpx.Client(base_url=self.base_url, timeout=10.0)

    def _json_response(self, response: httpx.Response) -> dict[str, Any]:
        return cast(dict[str, Any], response.json())

    def create_order(self, order_id: str, items: list[dict[str, Any]]) -> dict[str, Any]:
        with self._get_client() as client:
            response = client.post(
                protocols.ORDERS_ENDPOINT,
                json={
                    "order_id": order_id,
                    "items": items,
                },
            )
            response.raise_for_status()
            return self._json_response(response)

    def list_orders(self) -> dict[str, Any]:
        with self._get_client() as client:
            response = client.get(protocols.ORDERS_ENDPOINT)
            response.raise_for_status()
            return self._json_response(response)

    def get_order(self, order_id: str) -> dict[str, Any]:
        with self._get_client() as client:
            url = protocols.ORDER_ENDPOINT.format(order_id=order_id)
            response = client.get(url)
            response.raise_for_status()
            return self._json_response(response)

    def update_order(self, order_id: str, items: Optional[list[dict[str, Any]]] = None) -> dict[str, Any]:
        with self._get_client() as client:
            url = protocols.ORDER_ENDPOINT.format(order_id=order_id)
            json_data: dict[str, Any] = {}
            if items is not None:
                json_data["items"] = items
            response = client.put(url, json=json_data)
            response.raise_for_status()
            return self._json_response(response)

    def create_package(self, order_id: str, box_type_id: str) -> dict[str, Any]:
        with self._get_client() as client:
            url = protocols.PACKAGES_ENDPOINT.format(order_id=order_id)
            response = client.post(
                url,
                json={"box_type_id": box_type_id},
            )
            response.raise_for_status()
            return self._json_response(response)

    def list_packages(self, order_id: str) -> dict[str, Any]:
        with self._get_client() as client:
            url = protocols.PACKAGES_ENDPOINT.format(order_id=order_id)
            response = client.get(url)
            response.raise_for_status()
            return self._json_response(response)

    def add_item_to_package(self, package_id: str, product_id: str, quantity: int) -> dict[str, Any]:
        with self._get_client() as client:
            url = protocols.PACKAGE_ITEMS_ENDPOINT.format(package_id=package_id)
            response = client.post(
                url,
                json={
                    "product_id": product_id,
                    "quantity": quantity,
                },
            )
            response.raise_for_status()
            return self._json_response(response)

    def update_tracking(self, package_id: str, tracking_number: str) -> dict[str, Any]:
        with self._get_client() as client:
            url = protocols.PACKAGE_TRACKING_ENDPOINT.format(package_id=package_id)
            response = client.post(
                url,
                json={"tracking_number": tracking_number},
            )
            response.raise_for_status()
            return self._json_response(response)

    def get_packing_list_text(self, order_id: str) -> dict[str, Any]:
        with self._get_client() as client:
            url = protocols.PACKING_LIST_FORMAT_ENDPOINT.format(order_id=order_id)
            response = client.get(url)
            response.raise_for_status()
            return self._json_response(response)

    def get_packing_suggestions(self, order_id: str) -> dict[str, Any]:
        with self._get_client() as client:
            url = protocols.PACKING_SUGGESTIONS_ENDPOINT.format(order_id=order_id)
            response = client.get(url)
            response.raise_for_status()
            return self._json_response(response)

    def list_box_types(self) -> dict[str, Any]:
        with self._get_client() as client:
            response = client.get(protocols.BOX_TYPES_ENDPOINT)
            response.raise_for_status()
            return self._json_response(response)

    def get_material_stats(self) -> dict[str, Any]:
        with self._get_client() as client:
            response = client.get(protocols.MATERIAL_STATS_ENDPOINT)
            response.raise_for_status()
            return self._json_response(response)

    def save_template(self, order_id: str, template_name: str) -> dict[str, Any]:
        with self._get_client() as client:
            url = (
                protocols.PACKING_LISTS_ENDPOINT.format(order_id=order_id)
                + "/save-template"
            )
            response = client.post(
                url,
                json={"template_name": template_name},
            )
            response.raise_for_status()
            return self._json_response(response)

    def list_templates(self) -> dict[str, Any]:
        with self._get_client() as client:
            response = client.get(protocols.PACKING_TEMPLATES_ENDPOINT)
            response.raise_for_status()
            return self._json_response(response)

    def apply_template(self, order_id: str, template_id: str) -> dict[str, Any]:
        with self._get_client() as client:
            url = (
                protocols.PACKING_LISTS_ENDPOINT.format(order_id=order_id)
                + "/apply-template"
            )
            response = client.post(
                url,
                json={"template_id": template_id},
            )
            response.raise_for_status()
            return self._json_response(response)

    def health_check(self) -> dict[str, Any]:
        with self._get_client() as client:
            response = client.get("/health")
            response.raise_for_status()
            return self._json_response(response)
