package dto

// ApiResponse 统一标准响应结构体，与 FastAPI 返回的 {"code": 0, "message": "success", "data": ...} 保持 100% 契约一致
type ApiResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func Success[T any](data T) ApiResponse[T] {
	return ApiResponse[T]{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

func Fail(code int, message string) ApiResponse[any] {
	return ApiResponse[any]{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}
