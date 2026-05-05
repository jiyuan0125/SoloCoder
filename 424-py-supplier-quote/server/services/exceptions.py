from shared.constants.error_codes import ErrorCode


class BusinessException(Exception):
    def __init__(self, code: ErrorCode, message: str | None = None) -> None:
        self.code = code
        self.message = message or ""
        super().__init__(self.message or code.value)
