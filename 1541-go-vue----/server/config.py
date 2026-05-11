from pydantic_settings import BaseSettings
from pydantic import Field


class Settings(BaseSettings):
    app_name: str = "水利枢纽运行调度管理系统"
    version: str = "1.0.0"
    port: int = Field(default=8000, description="服务端口")
    host: str = Field(default="0.0.0.0", description="监听地址")
    database_url: str = Field(default="sqlite:///./reservoir.db", description="数据库连接")
    water_level_interval_minutes: int = Field(default=10, description="水位采集间隔(分钟)")
    gate_step_open_percent: int = Field(default=20, description="闸门开度步进(%)")
    gate_open_interval_minutes: int = Field(default=15, description="闸门开启间隔(分钟)")
    exit_flood_mode_hours: float = Field(default=24.0, description="退出防汛模式持续时间(小时)")
    red_alert_threshold: float = Field(default=0.95, description="红色预警阈值(设计洪水位百分比)")

    class Config:
        env_file = ".env"
        env_prefix = "RESERVOIR_"


settings = Settings()
