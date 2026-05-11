import os
import sys
import json
from typing import Optional
import click

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", ".."))

from src.client.api_client import APIClient


def get_client() -> APIClient:
    return APIClient()


@click.group()
@click.option("--server", default=None, help="服务端地址，如 http://localhost:8000")
@click.pass_context
def cli(ctx, server):
    """清洁生产审核管理系统命令行客户端"""
    if server:
        os.environ["SERVER_URL"] = server
    ctx.ensure_object(dict)
    ctx.obj["client"] = get_client()


@cli.command()
@click.pass_context
def health(ctx):
    """检查服务端健康状态"""
    client = ctx.obj["client"]
    try:
        result = client.health_check()
        click.echo(f"服务状态: {result['status']}")
    except Exception as e:
        click.echo(f"连接失败: {e}", err=True)
        sys.exit(1)


@cli.command("audit-create")
@click.option("--company", required=True, help="企业名称")
@click.option("--company-id", help="企业ID")
@click.option("--user-id", help="操作用户ID")
@click.option("--user-name", help="操作用户名称")
@click.pass_context
def audit_create(ctx, company, company_id, user_id, user_name):
    """创建新的审核项目"""
    client = ctx.obj["client"]
    try:
        result = client.create_audit(
            company_name=company,
            company_id=company_id,
            user_id=user_id,
            user_name=user_name,
        )
        click.echo(json.dumps(result, indent=2, ensure_ascii=False))
    except Exception as e:
        click.echo(f"创建失败: {e}", err=True)
        sys.exit(1)


@cli.command("audit-list")
@click.pass_context
def audit_list(ctx):
    """列出所有审核项目"""
    client = ctx.obj["client"]
    try:
        result = client.list_audits()
        if not result:
            click.echo("暂无审核项目")
            return
        for audit in result:
            click.echo(f"ID: {audit['id']} | 企业: {audit['company_name']} | 状态: {audit['status']}")
            if audit["is_overdue"]:
                click.echo("  ⚠️  已超期")
    except Exception as e:
        click.echo(f"查询失败: {e}", err=True)
        sys.exit(1)


@cli.command("audit-get")
@click.argument("audit_id", type=int)
@click.pass_context
def audit_get(ctx, audit_id):
    """获取审核项目详情"""
    client = ctx.obj["client"]
    try:
        result = client.get_audit(audit_id)
        click.echo(json.dumps(result, indent=2, ensure_ascii=False))
    except Exception as e:
        click.echo(f"查询失败: {e}", err=True)
        sys.exit(1)


@cli.command("audit-advance")
@click.argument("audit_id", type=int)
@click.option("--notes", help="阶段备注")
@click.option("--user-id", help="操作用户ID")
@click.option("--user-name", help="操作用户名称")
@click.pass_context
def audit_advance(ctx, audit_id, notes, user_id, user_name):
    """推进审核到下一阶段"""
    client = ctx.obj["client"]
    try:
        result = client.advance_stage(
            audit_id=audit_id,
            notes=notes,
            user_id=user_id,
            user_name=user_name,
        )
        click.echo(f"阶段完成: {result['stage_type']}")
    except Exception as e:
        click.echo(f"推进失败: {e}", err=True)
        sys.exit(1)


@cli.command("solution-create")
@click.argument("audit_id", type=int)
@click.option("--name", required=True, help="方案名称")
@click.option("--type", "solution_type", required=True, type=click.Choice(["no_low_cost", "medium_high_cost"]), help="方案类型")
@click.option("--description", help="方案描述")
@click.option("--expected-energy", type=float, help="预期节能量")
@click.option("--expected-investment", type=float, help="预期投资")
@click.option("--user-id", help="操作用户ID")
@click.option("--user-name", help="操作用户名称")
@click.pass_context
def solution_create(ctx, audit_id, name, solution_type, description, expected_energy, expected_investment, user_id, user_name):
    """创建方案"""
    client = ctx.obj["client"]
    try:
        result = client.create_solution(
            audit_id=audit_id,
            name=name,
            solution_type=solution_type,
            description=description,
            expected_energy_saving=expected_energy,
            expected_investment=expected_investment,
            user_id=user_id,
            user_name=user_name,
        )
        click.echo(json.dumps(result, indent=2, ensure_ascii=False))
    except Exception as e:
        click.echo(f"创建方案失败: {e}", err=True)
        sys.exit(1)


@cli.command("solution-add-dimensions")
@click.argument("solution_id", type=int)
@click.option("--dim1-name", required=True, help="维度1名称")
@click.option("--dim1-score", required=True, type=int, help="维度1评分(0-10)")
@click.option("--dim2-name", required=True, help="维度2名称")
@click.option("--dim2-score", required=True, type=int, help="维度2评分(0-10)")
@click.option("--dim3-name", required=True, help="维度3名称")
@click.option("--dim3-score", required=True, type=int, help="维度3评分(0-10)")
@click.option("--user-id", help="操作用户ID")
@click.option("--user-name", help="操作用户名称")
@click.pass_context
def solution_add_dimensions(ctx, solution_id, dim1_name, dim1_score, dim2_name, dim2_score, dim3_name, dim3_score, user_id, user_name):
    """为中高费方案添加维度评分(至少3个维度)"""
    client = ctx.obj["client"]
    dimensions = [
        {"dimension_name": dim1_name, "score": dim1_score},
        {"dimension_name": dim2_name, "score": dim2_score},
        {"dimension_name": dim3_name, "score": dim3_score},
    ]
    try:
        result = client.add_dimensions(
            solution_id=solution_id,
            dimensions=dimensions,
            user_id=user_id,
            user_name=user_name,
        )
        click.echo(json.dumps(result, indent=2, ensure_ascii=False))
    except Exception as e:
        click.echo(f"添加维度失败: {e}", err=True)
        sys.exit(1)


@cli.command("solution-filter")
@click.argument("solution_id", type=int)
@click.option("--user-id", help="操作用户ID")
@click.option("--user-name", help="操作用户名称")
@click.pass_context
def solution_filter(ctx, solution_id, user_id, user_name):
    """执行方案筛选"""
    client = ctx.obj["client"]
    try:
        result = client.filter_solution(
            solution_id=solution_id,
            user_id=user_id,
            user_name=user_name,
        )
        if result["passed"]:
            click.echo(f"✅ 方案通过筛选: {result['solution_name']}")
        else:
            click.echo(f"❌ 方案未通过筛选: {result['solution_name']}")
            if result.get("reason"):
                click.echo(f"   原因: {result['reason']}")
    except Exception as e:
        click.echo(f"筛选失败: {e}", err=True)
        sys.exit(1)


@cli.command("solution-implement")
@click.argument("solution_id", type=int)
@click.option("--actual-energy", required=True, type=float, help="实际节能量")
@click.option("--actual-investment", required=True, type=float, help="实际投资")
@click.option("--user-id", help="操作用户ID")
@click.option("--user-name", help="操作用户名称")
@click.pass_context
def solution_implement(ctx, solution_id, actual_energy, actual_investment, user_id, user_name):
    """记录方案实施结果"""
    client = ctx.obj["client"]
    try:
        result = client.implement_solution(
            solution_id=solution_id,
            actual_energy_saving=actual_energy,
            actual_investment=actual_investment,
            user_id=user_id,
            user_name=user_name,
        )
        click.echo(f"方案名称: {result['name']}")
        if result["is_effective"]:
            click.echo("✅ 实施达标")
        else:
            click.echo("❌ 实施不达标")
    except Exception as e:
        click.echo(f"记录实施失败: {e}", err=True)
        sys.exit(1)


@cli.command("audit-accept")
@click.argument("audit_id", type=int)
@click.option("--score", required=True, type=float, help="验收得分(0-100)")
@click.option("--notes", help="验收备注")
@click.option("--user-id", help="操作用户ID")
@click.option("--user-name", help="操作用户名称")
@click.pass_context
def audit_accept(ctx, audit_id, score, notes, user_id, user_name):
    """执行验收"""
    client = ctx.obj["client"]
    try:
        result = client.perform_acceptance(
            audit_id=audit_id,
            score=score,
            notes=notes,
            user_id=user_id,
            user_name=user_name,
        )
        click.echo(f"验收得分: {result['acceptance_score']}")
        click.echo(f"验收结果: {result['status']}")
        if result["is_overdue"]:
            click.echo("⚠️  项目超期")
    except Exception as e:
        click.echo(f"验收失败: {e}", err=True)
        sys.exit(1)


@cli.command("audit-summary")
@click.argument("audit_id", type=int)
@click.pass_context
def audit_summary(ctx, audit_id):
    """查看审核项目汇总"""
    client = ctx.obj["client"]
    try:
        result = client.get_audit_summary(audit_id)
        click.echo(f"总方案数: {result['total_solutions']}")
        click.echo(f"通过筛选: {result['passed_filter']}")
        click.echo(f"已实施: {result['implemented']}")
        click.echo(f"达标率: {result['effective_rate']:.2%}")
        click.echo(f"阶段进度: {result['stages_completed']}/{result['total_stages']}")
    except Exception as e:
        click.echo(f"查询汇总失败: {e}", err=True)
        sys.exit(1)


@cli.command("logs")
@click.option("--audit-id", type=int, help="指定审核项目ID")
@click.pass_context
def logs(ctx, audit_id):
    """查看审计日志"""
    client = ctx.obj["client"]
    try:
        result = client.list_logs(audit_id=audit_id)
        if not result:
            click.echo("暂无日志")
            return
        for log in result:
            click.echo(f"[{log['timestamp']}] {log['action']}")
            if log.get("user_name"):
                click.echo(f"  操作人: {log['user_name']}")
            if log.get("new_value"):
                click.echo(f"  内容: {log['new_value']}")
            click.echo("")
    except Exception as e:
        click.echo(f"查询日志失败: {e}", err=True)
        sys.exit(1)


if __name__ == "__main__":
    cli(obj={})
