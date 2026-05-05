from server.routers.supplier_router import router as supplier_router
from server.routers.purchase_router import router as purchase_router
from server.routers.quote_router import router as quote_router
from server.routers.order_router import router as order_router
from server.routers.report_router import router as report_router

__all__ = [
    "supplier_router",
    "purchase_router",
    "quote_router",
    "order_router",
    "report_router",
]
