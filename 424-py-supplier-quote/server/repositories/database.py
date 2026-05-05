from typing import Any, Type, TypeVar, cast

from server.repositories.base import BaseRepository
from server.repositories.supplier import SupplierRepository, SupplierQualificationReviewRepository
from server.repositories.purchase import PurchaseRepository
from server.repositories.quote import QuoteRepository
from server.repositories.order import OrderRepository

T = TypeVar("T", bound=BaseRepository[Any])


class RepositoryFactory:
    def __init__(self) -> None:
        self._repositories: dict[Type[Any], Any] = {}

    def get(self, repo_class: Type[T]) -> T:
        if repo_class not in self._repositories:
            self._repositories[repo_class] = repo_class()
        return cast(T, self._repositories[repo_class])

    def reset(self) -> None:
        self._repositories.clear()


_factory: RepositoryFactory | None = None


def initialize_repositories() -> None:
    global _factory
    _factory = RepositoryFactory()


def get_repository(repo_class: Type[T]) -> T:
    if _factory is None:
        initialize_repositories()
    assert _factory is not None
    return _factory.get(repo_class)
