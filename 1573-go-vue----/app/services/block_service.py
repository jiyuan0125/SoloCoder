from sqlalchemy.orm import Session
from app.models import BlockSection, BlockLog, Alert
from datetime import datetime
from typing import List, Optional


def get_block(db: Session, device_id: str) -> Optional[BlockSection]:
    return db.query(BlockSection).filter(BlockSection.device_id == device_id).first()


def get_all_blocks(db: Session) -> List[BlockSection]:
    return db.query(BlockSection).all()


def create_block(db: Session, device_id: str, name: str) -> BlockSection:
    db_block = BlockSection(device_id=device_id, name=name)
    db.add(db_block)
    db.commit()
    db.refresh(db_block)
    return db_block


def occupy_block(db: Session, device_id: str) -> dict:
    block = get_block(db, device_id)
    if not block:
        return {"success": False, "message": "闭塞分区不存在"}
    
    if not block.circuit_ok:
        return {"success": False, "message": "轨道电路故障，无法占用"}
    
    if block.mode == "automatic":
        blocks = get_all_blocks(db)
        occupied_count = sum(1 for b in blocks if b.occupied)
        if occupied_count >= 2:
            return {"success": False, "message": "前方闭塞分区数量不足"}
    
    block.occupied = True
    block.last_updated = datetime.utcnow()
    
    log = BlockLog(
        section_id=device_id,
        action="occupied",
        occupied=True
    )
    db.add(log)
    db.commit()
    db.refresh(block)
    
    return {"success": True, "message": "闭塞分区已占用", "block": block}


def release_block(db: Session, device_id: str) -> dict:
    block = get_block(db, device_id)
    if not block:
        return {"success": False, "message": "闭塞分区不存在"}
    
    block.occupied = False
    block.last_updated = datetime.utcnow()
    
    log = BlockLog(
        section_id=device_id,
        action="released",
        occupied=False
    )
    db.add(log)
    db.commit()
    db.refresh(block)
    
    return {"success": True, "message": "闭塞分区已释放", "block": block}


def set_circuit_status(db: Session, device_id: str, circuit_ok: bool) -> dict:
    block = get_block(db, device_id)
    if not block:
        return {"success": False, "message": "闭塞分区不存在"}
    
    block.circuit_ok = circuit_ok
    if not circuit_ok:
        block.mode = "semi-automatic"
        alert = Alert(
            device_type="block",
            device_id=device_id,
            message="轨道电路故障，已降级为半自动闭塞"
        )
        db.add(alert)
    else:
        block.mode = "automatic"
    
    block.last_updated = datetime.utcnow()
    db.commit()
    db.refresh(block)
    
    return {"success": True, "message": "轨道电路状态已更新", "block": block}
