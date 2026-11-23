# -*- coding: utf-8 -*-
"""后端全局系统配置。"""
from __future__ import annotations

from pathlib import Path
from pydantic_settings import BaseSettings, SettingsConfigDict


class AppSettings(BaseSettings):
    app_name: str = "周期罗盘 CycleCompass API"
    app_version: str = "0.1.0"
    debug: bool = False
    host: str = "0.0.0.0"
    port: int = 8000
    cors_origins: list[str] = ["*"]
    database_url: str = "sqlite+aiosqlite:///./cycle_compass.db"
    redis_url: str = "redis://localhost:6379/0"

    model_config = SettingsConfigDict(env_prefix="CYCLE_", env_file=".env", extra="ignore")


settings = AppSettings()
