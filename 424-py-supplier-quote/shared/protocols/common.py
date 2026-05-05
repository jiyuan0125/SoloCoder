from typing import Generic, Optional, TypeVar, cast

from pydantic import Field

from shared.models.base import BaseModel
from shared.constants.error_codes import ErrorCode, ERROR_MESSAGES

T = TypeVar("T")


class ApiResponse(BaseModel, Generic[T]):
    is_success: bool = Field(..., description="是否成功")
    code: ErrorCode = Field(default=ErrorCode.SUCCESS, description="错误码")
    message: str = Field(default="", description="消息")
    data: Optional[T] = Field(default=None, description="数据")

    @staticmethod
    def create_success(data: T, message: str = "操作成功") -> "ApiResponse[T]":
        return ApiResponse[T](
            is_success=True, code=ErrorCode.SUCCESS, message=message, data=data
        )

    @staticmethod
    def create_error(code: ErrorCode, message: Optional[str] = None) -> "ApiResponse[T]":
        return cast(
            ApiResponse[T],
            ApiResponse[None](
                is_success=False,
                code=code,
                message=message or ERROR_MESSAGES.get(code, "未知错误"),
                data=None,
            ),
        )


class PaginationParams(BaseModel):
    page: int = Field(default=1, ge=1, description="页码")
    page_size: int = Field(default=20, ge=1, le=100, description="每页条数")


class PaginatedResponse(BaseModel, Generic[T]):
    items: list[T] = Field(..., description="数据列表")
    total: int = Field(..., ge=0, description="总条数")
    page: int = Field(..., ge=1, description="当前页码")
    page_size: int = Field(..., ge=1, description="每页条数")
    total_pages: int = Field(..., ge=0, description="总页数")
