# -*- coding: utf-8 -*-
"""统一 API 响应包装协议。"""
from __future__ import annotations

from typing import Generic, Optional, TypeVar
from pydantic import BaseModel

T = TypeVar("T")


class ApiResponse(BaseModel, Generic[T]):
    code: int = 0
    message: str = "success"
    data: Optional[T] = None

    @classmethod
    def success(cls, data: T, message: str = "success") -> ApiResponse[T]:
        return cls(code=0, message=message, data=data)

    @classmethod
    def fail(cls, message: str, code: int = 1) -> ApiResponse[None]:
        return cls(code=code, message=message, data=None)
