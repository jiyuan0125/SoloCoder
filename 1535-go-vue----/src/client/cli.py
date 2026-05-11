import json
import click
from .api_client import ApiClient


def format_json(obj, indent: int = 2) -> str:
    return json.dumps(obj, ensure_ascii=False, indent=indent)


@click.group()
@click.option("--server", envvar="APP_SERVER_URL", default="http://localhost:8000", help="服务端 URL")
@click.pass_context
def main(ctx, server):
    """环评全流程管理系统 - 命令行客户端"""
    ctx.ensure_object(dict)
    ctx.obj["client"] = ApiClient(base_url=server)


@main.group()
def project():
    """项目管理命令"""
    pass


@project.command(name="create")
@click.option("--name", "-n", required=True, help="项目名称")
@click.option("--company", "-c", required=True, help="委托单位")
@click.option("--description", "-d", required=True, help="项目描述")
@click.pass_context
def project_create(ctx, name, company, description):
    """创建新项目"""
    client = ctx.obj["client"]
    result = client.create_project(name, company, description)
    click.echo(format_json(result))


@project.command(name="list")
@click.option("--status", "-s", default=None, help="按状态过滤")
@click.pass_context
def project_list(ctx, status):
    """列出所有项目"""
    client = ctx.obj["client"]
    projects = client.list_projects(status)
    click.echo(format_json(projects))


@project.command(name="get")
@click.argument("project_id")
@click.pass_context
def project_get(ctx, project_id):
    """查看项目详情"""
    client = ctx.obj["client"]
    project = client.get_project(project_id)
    click.echo(format_json(project))


@project.command(name="update")
@click.argument("project_id")
@click.option("--name", "-n", default=None, help="新名称")
@click.option("--company", "-c", default=None, help="新委托单位")
@click.option("--description", "-d", default=None, help="新描述")
@click.pass_context
def project_update(ctx, project_id, name, company, description):
    """更新项目信息"""
    client = ctx.obj["client"]
    result = client.update_project(project_id, name, company, description)
    click.echo(format_json(result))


@project.command(name="start-preparation")
@click.argument("project_id")
@click.pass_context
def project_start_preparation(ctx, project_id):
    """启动环评编制阶段"""
    client = ctx.obj["client"]
    result = client.start_preparation(project_id)
    click.echo(format_json(result))


@project.command(name="complete-preparation")
@click.argument("project_id")
@click.pass_context
def project_complete_preparation(ctx, project_id):
    """完成环评编制阶段"""
    client = ctx.obj["client"]
    result = client.complete_preparation(project_id)
    click.echo(format_json(result))


@project.command(name="submit-evaluation")
@click.argument("project_id")
@click.option("--compliance", required=True, type=float, help="合规性评分 (0-100)")
@click.option("--technology", required=True, type=float, help="技术可行性评分 (0-100)")
@click.option("--environmental", required=True, type=float, help="环境影响评分 (0-100)")
@click.option("--feasibility", required=True, type=float, help="经济可行性评分 (0-100)")
@click.pass_context
def project_submit_evaluation(ctx, project_id, compliance, technology, environmental, feasibility):
    """提交评估打分（四项评分，低于60分退回）"""
    client = ctx.obj["client"]
    result = client.submit_evaluation(project_id, compliance, technology, environmental, feasibility)
    click.echo(format_json(result))


@project.command(name="start-publicity")
@click.argument("project_id")
@click.pass_context
def project_start_publicity(ctx, project_id):
    """启动审批公示阶段（7个工作日）"""
    client = ctx.obj["client"]
    result = client.start_publicity(project_id)
    click.echo(format_json(result))


@project.command(name="add-opinion")
@click.argument("project_id")
@click.option("--content", "-c", required=True, help="公众意见内容")
@click.pass_context
def project_add_opinion(ctx, project_id, content):
    """添加公众意见（需在10个工作日内回复）"""
    client = ctx.obj["client"]
    result = client.add_public_opinion(project_id, content)
    click.echo(format_json(result))


@project.command(name="respond-opinion")
@click.argument("project_id")
@click.option("--opinion-id", required=True, help="意见ID")
@click.option("--response", "-r", required=True, help="回复内容")
@click.pass_context
def project_respond_opinion(ctx, project_id, opinion_id, response):
    """回复公众意见"""
    client = ctx.obj["client"]
    result = client.respond_to_opinion(project_id, opinion_id, response)
    click.echo(format_json(result))


@project.command(name="complete-publicity")
@click.argument("project_id")
@click.pass_context
def project_complete_publicity(ctx, project_id):
    """完成审批公示阶段（需所有意见已回复）"""
    client = ctx.obj["client"]
    result = client.complete_publicity(project_id)
    click.echo(format_json(result))


@project.command(name="approve")
@click.argument("project_id")
@click.option("--decision", "-d", required=True, help="审批意见")
@click.pass_context
def project_approve(ctx, project_id, decision):
    """审批项目（最终批复）"""
    client = ctx.obj["client"]
    result = client.approve_project(project_id, decision)
    click.echo(format_json(result))


@project.command(name="resubmit")
@click.argument("project_id")
@click.pass_context
def project_resubmit(ctx, project_id):
    """修改后重新提交评估（仅退回项目可用）"""
    client = ctx.obj["client"]
    result = client.resubmit_project(project_id)
    click.echo(format_json(result))


@main.group()
def todo():
    """待办事项管理命令"""
    pass


@todo.command(name="list")
@click.option("--completed", is_flag=True, default=None, help="只显示已完成")
@click.option("--pending", is_flag=True, default=None, help="只显示待处理")
@click.pass_context
def todo_list(ctx, completed, pending):
    """列出待办事项"""
    client = ctx.obj["client"]
    if pending:
        todos = client.list_todos(completed=False)
    elif completed:
        todos = client.list_todos(completed=True)
    else:
        todos = client.list_todos()
    click.echo(format_json(todos))


@todo.command(name="get")
@click.argument("todo_id")
@click.pass_context
def todo_get(ctx, todo_id):
    """查看待办详情"""
    client = ctx.obj["client"]
    todo = client.get_todo(todo_id)
    click.echo(format_json(todo))


@todo.command(name="complete")
@click.argument("todo_id")
@click.pass_context
def todo_complete(ctx, todo_id):
    """标记待办为已完成"""
    client = ctx.obj["client"]
    result = client.complete_todo(todo_id)
    click.echo(format_json(result))


@main.command(name="info")
@click.pass_context
def server_info(ctx):
    """查看服务端信息"""
    client = ctx.obj["client"]
    info = client.get_root()
    click.echo(format_json(info))


if __name__ == "__main__":
    main(obj={})
