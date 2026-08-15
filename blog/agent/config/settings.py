# blog/agent/config/settings.py
from pydantic import model_validator
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """笔墨精灵智能体服务配置"""

    # DeepSeek API
    deepseek_api_key: str = ""
    deepseek_base_url: str = "https://api.deepseek.com"
    deepseek_model: str = "deepseek-chat"

    # Redis 短期记忆
    redis_url: str = "redis://localhost:6379"

    # Agent 配置
    short_memory_ttl_seconds: int = 1800  # 30 分钟
    max_history_rounds: int = 5
    max_retrieval_iterations: int = 8

    # 日志
    log_level: str = "INFO"

    model_config = {"env_file": ".env", "env_prefix": "", "extra": "ignore"}

    @model_validator(mode="after")
    def _require_api_key(self) -> "Settings":
        """缺 API key 时启动即失败，避免带病上线"""
        if not self.deepseek_api_key:
            raise ValueError("DEEPSEEK_API_KEY 未配置，无法启动笔墨精灵")
        return self


# 全局单例
settings = Settings()
