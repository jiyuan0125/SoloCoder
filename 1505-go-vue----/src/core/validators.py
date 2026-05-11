from typing import List

from .models import (
    BatchCreate,
    Batch,
    RecipeCreate,
    QualityInspectionCreate,
    ValidationError,
    InspectionResult,
)


def validate_recipe(recipe: RecipeCreate) -> None:
    if not recipe.name or not recipe.name.strip():
        raise ValidationError("配方名称不能为空")
    if not recipe.product_category or not recipe.product_category.strip():
        raise ValidationError("产品类别不能为空")
    if len(recipe.ingredients) < 2:
        raise ValidationError("配方原材料清单至少需要两条")
    
    material_names = set()
    for ing in recipe.ingredients:
        if not ing.material_name or not ing.material_name.strip():
            raise ValidationError("原材料名称不能为空")
        if ing.quantity <= 0:
            raise ValidationError(f"原材料 {ing.material_name} 用量必须大于0")
        if ing.material_name in material_names:
            raise ValidationError(f"原材料 {ing.material_name} 重复")
        material_names.add(ing.material_name)


def validate_batch(batch: BatchCreate) -> None:
    if batch.planned_quantity <= 0:
        raise ValidationError("计划数量必须大于0")
    if batch.actual_quantity is not None:
        if batch.actual_quantity < 0:
            raise ValidationError("实际数量不能为负数")
        if batch.actual_quantity > batch.planned_quantity * 1.1:
            raise ValidationError("实际成品数量不能超过计划的110%")


def validate_batch_actual_quantity(
    actual_quantity: float,
    planned_quantity: float,
) -> None:
    if actual_quantity < 0:
        raise ValidationError("实际数量不能为负数")
    if actual_quantity > planned_quantity * 1.1:
        raise ValidationError("实际成品数量不能超过计划的110%")


def validate_quality_inspection(
    inspection: QualityInspectionCreate,
    batch: Batch,
    existing_inspections: List,
) -> None:
    for existing in existing_inspections:
        if existing.result == InspectionResult.PASS:
            raise ValidationError("已有合格质检的批次不能再加新质检记录")
    
    if inspection.result == InspectionResult.PASS and inspection.inspection_value is None:
        raise ValidationError("检测数值为空但结果填合格的是矛盾数据要拒绝")
