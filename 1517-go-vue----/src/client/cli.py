import argparse
import sys
import json
from datetime import date, datetime
from typing import Optional

from .api_client import APIClient


def print_json(data):
    print(json.dumps(data, ensure_ascii=False, indent=2, default=str))


def parse_date(value: str) -> date:
    return datetime.strptime(value, "%Y-%m-%d").date()


def parse_datetime(value: str) -> datetime:
    try:
        return datetime.strptime(value, "%Y-%m-%d %H:%M")
    except ValueError:
        return datetime.strptime(value, "%Y-%m-%dT%H:%M")


class CLI:
    def __init__(self):
        self.client = APIClient()

    def close(self):
        self.client.close()

    def cmd_health(self, args):
        if self.client.health_check():
            print("服务器运行正常")
        else:
            print("无法连接服务器")
            sys.exit(1)

    def cmd_animal_create(self, args):
        data = {
            "name": args.name,
            "species": args.species,
            "breed": args.breed,
            "health_status": args.health_status,
            "cage": args.cage,
            "arrival_date": args.arrival_date or date.today().isoformat(),
            "is_adoptable": args.is_adoptable,
        }
        if args.description:
            data["description"] = args.description
        result = self.client.create_animal(data)
        print_json(result)

    def cmd_animal_list(self, args):
        result = self.client.list_animals(
            species=args.species,
            health_status=args.health_status,
            is_adoptable=args.is_adoptable
        )
        print_json(result)

    def cmd_animal_get(self, args):
        result = self.client.get_animal(args.id)
        print_json(result)

    def cmd_animal_update(self, args):
        data = {}
        if args.name is not None:
            data["name"] = args.name
        if args.species is not None:
            data["species"] = args.species
        if args.breed is not None:
            data["breed"] = args.breed
        if args.health_status is not None:
            data["health_status"] = args.health_status
        if args.cage is not None:
            data["cage"] = args.cage
        if args.arrival_date is not None:
            data["arrival_date"] = args.arrival_date
        if args.description is not None:
            data["description"] = args.description
        if args.is_adoptable is not None:
            data["is_adoptable"] = args.is_adoptable
        if not data:
            print("请提供至少一个要更新的字段")
            sys.exit(1)
        result = self.client.update_animal(args.id, data)
        print_json(result)

    def cmd_animal_delete(self, args):
        result = self.client.delete_animal(args.id)
        print_json(result)

    def cmd_adopter_create(self, args):
        data = {
            "name": args.name,
            "phone": args.phone,
            "address": args.address,
            "has_bad_record": args.has_bad_record,
        }
        result = self.client.create_adopter(data)
        print_json(result)

    def cmd_adopter_list(self, args):
        result = self.client.list_adopters(
            has_bad_record=args.has_bad_record
        )
        print_json(result)

    def cmd_adopter_get(self, args):
        result = self.client.get_adopter(args.id)
        print_json(result)

    def cmd_adopter_update(self, args):
        data = {}
        if args.name is not None:
            data["name"] = args.name
        if args.phone is not None:
            data["phone"] = args.phone
        if args.address is not None:
            data["address"] = args.address
        if args.has_bad_record is not None:
            data["has_bad_record"] = args.has_bad_record
        if not data:
            print("请提供至少一个要更新的字段")
            sys.exit(1)
        result = self.client.update_adopter(args.id, data)
        print_json(result)

    def cmd_adopter_delete(self, args):
        result = self.client.delete_adopter(args.id)
        print_json(result)

    def cmd_adoption_create(self, args):
        data = {
            "animal_id": args.animal_id,
            "adopter_id": args.adopter_id,
        }
        if args.notes:
            data["notes"] = args.notes
        result = self.client.create_adoption(data)
        print_json(result)

    def cmd_adoption_list(self, args):
        result = self.client.list_adoptions(
            animal_id=args.animal_id,
            adopter_id=args.adopter_id,
            status=args.status
        )
        print_json(result)

    def cmd_adoption_get(self, args):
        result = self.client.get_adoption(args.id)
        print_json(result)

    def cmd_adoption_approve(self, args):
        if args.extra:
            result = self.client.approve_adoption_extra(args.id)
        else:
            result = self.client.approve_adoption(args.id)
        print_json(result)

    def cmd_adoption_reject(self, args):
        result = self.client.reject_adoption(args.id)
        print_json(result)

    def cmd_adoption_cancel(self, args):
        result = self.client.cancel_adoption(args.id)
        print_json(result)

    def cmd_adoption_finalize(self, args):
        result = self.client.finalize_adoption(args.id)
        print_json(result)

    def cmd_followup_list(self, args):
        result = self.client.list_follow_ups(
            adoption_id=args.adoption_id,
            status=args.status,
            adopter_id=args.adopter_id
        )
        print_json(result)

    def cmd_followup_get(self, args):
        result = self.client.get_follow_up(args.id)
        print_json(result)

    def cmd_followup_complete(self, args):
        result = self.client.complete_follow_up(args.id, args.notes)
        print_json(result)

    def cmd_followup_check_overdue(self, args):
        result = self.client.check_overdue_follow_ups()
        print_json(result)

    def cmd_donation_create(self, args):
        data = {
            "donor_name": args.donor_name,
            "donation_type": args.donation_type,
            "amount": args.amount or 0,
        }
        if args.donor_phone:
            data["donor_phone"] = args.donor_phone
        if args.description:
            data["description"] = args.description
        result = self.client.create_donation(data)
        print_json(result)

    def cmd_donation_list(self, args):
        result = self.client.list_donations(
            donation_type=args.donation_type
        )
        print_json(result)

    def cmd_donation_stats(self, args):
        result = self.client.get_donation_stats()
        print_json(result)

    def cmd_donation_get(self, args):
        result = self.client.get_donation(args.id)
        print_json(result)

    def cmd_appointment_create(self, args):
        data = {
            "adoption_id": args.adoption_id,
            "scheduled_time": args.scheduled_time,
            "location": args.location,
        }
        if args.notes:
            data["notes"] = args.notes
        result = self.client.create_appointment(data)
        print_json(result)

    def cmd_appointment_list(self, args):
        result = self.client.list_appointments(
            adoption_id=args.adoption_id,
            status=args.status
        )
        print_json(result)

    def cmd_appointment_get(self, args):
        result = self.client.get_appointment(args.id)
        print_json(result)

    def cmd_appointment_complete(self, args):
        result = self.client.complete_appointment(args.id, args.passed, args.notes)
        print_json(result)

    def cmd_appointment_cancel(self, args):
        result = self.client.cancel_appointment(args.id)
        print_json(result)


def create_parser():
    parser = argparse.ArgumentParser(
        prog="rescue-cli",
        description="动物救助站管理系统命令行客户端"
    )
    subparsers = parser.add_subparsers(dest="command", help="可用命令")

    subparsers.add_parser("health", help="检查服务器健康状态")

    animal_parser = subparsers.add_parser("animal", help="动物管理")
    animal_sub = animal_parser.add_subparsers(dest="animal_cmd", help="动物管理命令")

    create_animal = animal_sub.add_parser("create", help="创建动物档案")
    create_animal.add_argument("--name", required=True, help="动物名称")
    create_animal.add_argument("--species", required=True, choices=["dog", "cat", "bird", "rabbit", "other"], help="物种")
    create_animal.add_argument("--breed", required=True, help="品种")
    create_animal.add_argument("--health-status", required=True, choices=["healthy", "isolation", "treatment", "recovering"], help="健康状态")
    create_animal.add_argument("--cage", required=True, help="笼舍编号")
    create_animal.add_argument("--arrival-date", help="收容日期 (YYYY-MM-DD)")
    create_animal.add_argument("--description", help="描述")
    create_animal.add_argument("--is-adoptable", action="store_true", default=True, help="是否可领养")

    list_animal = animal_sub.add_parser("list", help="列出动物")
    list_animal.add_argument("--species", choices=["dog", "cat", "bird", "rabbit", "other"], help="按物种过滤")
    list_animal.add_argument("--health-status", choices=["healthy", "isolation", "treatment", "recovering"], help="按健康状态过滤")
    list_animal.add_argument("--is-adoptable", action="store_true", default=None, help="只显示可领养")
    list_animal.add_argument("--not-adoptable", dest="is_adoptable", action="store_false", help="只显示不可领养")

    get_animal = animal_sub.add_parser("get", help="获取动物详情")
    get_animal.add_argument("--id", required=True, help="动物ID")

    update_animal = animal_sub.add_parser("update", help="更新动物信息")
    update_animal.add_argument("--id", required=True, help="动物ID")
    update_animal.add_argument("--name", help="动物名称")
    update_animal.add_argument("--species", choices=["dog", "cat", "bird", "rabbit", "other"], help="物种")
    update_animal.add_argument("--breed", help="品种")
    update_animal.add_argument("--health-status", choices=["healthy", "isolation", "treatment", "recovering"], help="健康状态")
    update_animal.add_argument("--cage", help="笼舍编号")
    update_animal.add_argument("--arrival-date", help="收容日期")
    update_animal.add_argument("--description", help="描述")
    group_adopt = update_animal.add_mutually_exclusive_group()
    group_adopt.add_argument("--is-adoptable", dest="is_adoptable", action="store_true", default=None)
    group_adopt.add_argument("--not-adoptable", dest="is_adoptable", action="store_false")

    del_animal = animal_sub.add_parser("delete", help="删除动物")
    del_animal.add_argument("--id", required=True, help="动物ID")

    adopter_parser = subparsers.add_parser("adopter", help="领养人管理")
    adopter_sub = adopter_parser.add_subparsers(dest="adopter_cmd", help="领养人管理命令")

    create_adopter = adopter_sub.add_parser("create", help="创建领养人档案")
    create_adopter.add_argument("--name", required=True, help="姓名")
    create_adopter.add_argument("--phone", required=True, help="电话")
    create_adopter.add_argument("--address", required=True, help="地址")
    create_adopter.add_argument("--has-bad-record", action="store_true", help="有不良记录")

    list_adopter = adopter_sub.add_parser("list", help="列出领养人")
    list_adopter.add_argument("--has-bad-record", action="store_true", default=None, help="只显示有不良记录")
    list_adopter.add_argument("--no-bad-record", dest="has_bad_record", action="store_false", help="只显示无不良记录")

    get_adopter = adopter_sub.add_parser("get", help="获取领养人详情")
    get_adopter.add_argument("--id", required=True, help="领养人ID")

    update_adopter = adopter_sub.add_parser("update", help="更新领养人信息")
    update_adopter.add_argument("--id", required=True, help="领养人ID")
    update_adopter.add_argument("--name", help="姓名")
    update_adopter.add_argument("--phone", help="电话")
    update_adopter.add_argument("--address", help="地址")
    group_bad = update_adopter.add_mutually_exclusive_group()
    group_bad.add_argument("--has-bad-record", dest="has_bad_record", action="store_true", default=None)
    group_bad.add_argument("--no-bad-record", dest="has_bad_record", action="store_false")

    del_adopter = adopter_sub.add_parser("delete", help="删除领养人")
    del_adopter.add_argument("--id", required=True, help="领养人ID")

    adoption_parser = subparsers.add_parser("adoption", help="领养申请管理")
    adoption_sub = adoption_parser.add_subparsers(dest="adoption_cmd", help="领养申请命令")

    create_adoption = adoption_sub.add_parser("create", help="创建领养申请")
    create_adoption.add_argument("--animal-id", required=True, help="动物ID")
    create_adoption.add_argument("--adopter-id", required=True, help="领养人ID")
    create_adoption.add_argument("--notes", help="备注")

    list_adoption = adoption_sub.add_parser("list", help="列出领养申请")
    list_adoption.add_argument("--animal-id", help="按动物过滤")
    list_adoption.add_argument("--adopter-id", help="按领养人过滤")
    list_adoption.add_argument("--status", help="按状态过滤")

    get_adoption = adoption_sub.add_parser("get", help="获取申请详情")
    get_adoption.add_argument("--id", required=True, help="申请ID")

    approve_adoption = adoption_sub.add_parser("approve", help="批准申请")
    approve_adoption.add_argument("--id", required=True, help="申请ID")
    approve_adoption.add_argument("--extra", action="store_true", help="额外审批通过")

    reject_adoption = adoption_sub.add_parser("reject", help="拒绝申请")
    reject_adoption.add_argument("--id", required=True, help="申请ID")

    cancel_adoption = adoption_sub.add_parser("cancel", help="取消申请")
    cancel_adoption.add_argument("--id", required=True, help="申请ID")

    finalize_adoption = adoption_sub.add_parser("finalize", help="完成领养")
    finalize_adoption.add_argument("--id", required=True, help="申请ID")

    followup_parser = subparsers.add_parser("followup", help="回访管理")
    followup_sub = followup_parser.add_subparsers(dest="followup_cmd", help="回访命令")

    list_followup = followup_sub.add_parser("list", help="列出来访")
    list_followup.add_argument("--adoption-id", help="按领养申请过滤")
    list_followup.add_argument("--adopter-id", help="按领养人过滤")
    list_followup.add_argument("--status", choices=["pending", "completed", "overdue"], help="按状态过滤")

    get_followup = followup_sub.add_parser("get", help="获取回访详情")
    get_followup.add_argument("--id", required=True, help="回访ID")

    complete_followup = followup_sub.add_parser("complete", help="完成回访")
    complete_followup.add_argument("--id", required=True, help="回访ID")
    complete_followup.add_argument("--notes", help="回访记录")

    followup_sub.add_parser("check-overdue", help="检查逾期回访")

    donation_parser = subparsers.add_parser("donation", help="捐赠管理")
    donation_sub = donation_parser.add_subparsers(dest="donation_cmd", help="捐赠命令")

    create_donation = donation_sub.add_parser("create", help="创建捐赠记录")
    create_donation.add_argument("--donor-name", required=True, help="捐赠人姓名")
    create_donation.add_argument("--donor-phone", help="捐赠人电话")
    create_donation.add_argument("--donation-type", required=True, choices=["money", "food", "supplies", "other"], help="捐赠类型")
    create_donation.add_argument("--amount", type=float, default=0, help="金额（资金捐赠必填）")
    create_donation.add_argument("--description", help="描述")

    list_donation = donation_sub.add_parser("list", help="列出捐赠")
    list_donation.add_argument("--donation-type", choices=["money", "food", "supplies", "other"], help="按类型过滤")

    donation_sub.add_parser("stats", help="捐赠统计")

    get_donation = donation_sub.add_parser("get", help="获取捐赠详情")
    get_donation.add_argument("--id", required=True, help="捐赠ID")

    appointment_parser = subparsers.add_parser("appointment", help="预约管理")
    appointment_sub = appointment_parser.add_subparsers(dest="appointment_cmd", help="预约命令")

    create_appointment = appointment_sub.add_parser("create", help="创建预约")
    create_appointment.add_argument("--adoption-id", required=True, help="领养申请ID")
    create_appointment.add_argument("--scheduled-time", required=True, help="预约时间 (YYYY-MM-DD HH:MM)")
    create_appointment.add_argument("--location", required=True, help="预约地点")
    create_appointment.add_argument("--notes", help="备注")

    list_appointment = appointment_sub.add_parser("list", help="列出预约")
    list_appointment.add_argument("--adoption-id", help="按领养申请过滤")
    list_appointment.add_argument("--status", choices=["scheduled", "completed", "cancelled"], help="按状态过滤")

    get_appointment = appointment_sub.add_parser("get", help="获取预约详情")
    get_appointment.add_argument("--id", required=True, help="预约ID")

    complete_appointment = appointment_sub.add_parser("complete", help="完成预约")
    complete_appointment.add_argument("--id", required=True, help="预约ID")
    complete_appointment.add_argument("--passed", action="store_true", help="面谈通过")
    complete_appointment.add_argument("--not-passed", dest="passed", action="store_false", help="面谈未通过")
    complete_appointment.add_argument("--notes", help="备注")
    complete_appointment.set_defaults(passed=True)

    cancel_appointment = appointment_sub.add_parser("cancel", help="取消预约")
    cancel_appointment.add_argument("--id", required=True, help="预约ID")

    return parser


def main():
    parser = create_parser()
    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        sys.exit(0)

    cli = CLI()

    try:
        if args.command == "health":
            cli.cmd_health(args)
        elif args.command == "animal":
            if args.animal_cmd == "create":
                cli.cmd_animal_create(args)
            elif args.animal_cmd == "list":
                cli.cmd_animal_list(args)
            elif args.animal_cmd == "get":
                cli.cmd_animal_get(args)
            elif args.animal_cmd == "update":
                cli.cmd_animal_update(args)
            elif args.animal_cmd == "delete":
                cli.cmd_animal_delete(args)
            else:
                print("请指定动物管理子命令")
                sys.exit(1)
        elif args.command == "adopter":
            if args.adopter_cmd == "create":
                cli.cmd_adopter_create(args)
            elif args.adopter_cmd == "list":
                cli.cmd_adopter_list(args)
            elif args.adopter_cmd == "get":
                cli.cmd_adopter_get(args)
            elif args.adopter_cmd == "update":
                cli.cmd_adopter_update(args)
            elif args.adopter_cmd == "delete":
                cli.cmd_adopter_delete(args)
            else:
                print("请指定领养人管理子命令")
                sys.exit(1)
        elif args.command == "adoption":
            if args.adoption_cmd == "create":
                cli.cmd_adoption_create(args)
            elif args.adoption_cmd == "list":
                cli.cmd_adoption_list(args)
            elif args.adoption_cmd == "get":
                cli.cmd_adoption_get(args)
            elif args.adoption_cmd == "approve":
                cli.cmd_adoption_approve(args)
            elif args.adoption_cmd == "reject":
                cli.cmd_adoption_reject(args)
            elif args.adoption_cmd == "cancel":
                cli.cmd_adoption_cancel(args)
            elif args.adoption_cmd == "finalize":
                cli.cmd_adoption_finalize(args)
            else:
                print("请指定领养管理子命令")
                sys.exit(1)
        elif args.command == "followup":
            if args.followup_cmd == "list":
                cli.cmd_followup_list(args)
            elif args.followup_cmd == "get":
                cli.cmd_followup_get(args)
            elif args.followup_cmd == "complete":
                cli.cmd_followup_complete(args)
            elif args.followup_cmd == "check-overdue":
                cli.cmd_followup_check_overdue(args)
            else:
                print("请指定回访管理子命令")
                sys.exit(1)
        elif args.command == "donation":
            if args.donation_cmd == "create":
                cli.cmd_donation_create(args)
            elif args.donation_cmd == "list":
                cli.cmd_donation_list(args)
            elif args.donation_cmd == "stats":
                cli.cmd_donation_stats(args)
            elif args.donation_cmd == "get":
                cli.cmd_donation_get(args)
            else:
                print("请指定捐赠管理子命令")
                sys.exit(1)
        elif args.command == "appointment":
            if args.appointment_cmd == "create":
                cli.cmd_appointment_create(args)
            elif args.appointment_cmd == "list":
                cli.cmd_appointment_list(args)
            elif args.appointment_cmd == "get":
                cli.cmd_appointment_get(args)
            elif args.appointment_cmd == "complete":
                cli.cmd_appointment_complete(args)
            elif args.appointment_cmd == "cancel":
                cli.cmd_appointment_cancel(args)
            else:
                print("请指定预约管理子命令")
                sys.exit(1)
    except Exception as e:
        print(f"错误: {e}")
        sys.exit(1)
    finally:
        cli.close()


if __name__ == "__main__":
    main()
