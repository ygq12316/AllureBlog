from pydantic import model_validator
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """笔墨精灵智能体服务配置"""

    # DeepSeek API
    deepseek_api_key: str = ""
    deepseek_base_url: str = "https://api.deepseek.com"
    deepseek_model: str = "deepseek-chat"

    # Redis（多轮记忆 + 每日计数）
    redis_url: str = "redis://localhost:6379"

    # Go 博客后端（登录校验 + 标题检索）
    blog_api_base: str = "http://localhost:8080"

    # Agent 行为
    short_memory_ttl_seconds: int = 1800  # 历史 30 分钟无对话即遗忘
    max_history_rounds: int = 5           # 记忆窗口（轮）
    max_tool_loops: int = 2               # 单轮对话最多几次工具调用机会
    daily_message_limit: int = 10         # 每访客每日消息数

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
