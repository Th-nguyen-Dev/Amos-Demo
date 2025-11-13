"""Configuration management using Pydantic Settings."""
from pydantic_settings import BaseSettings, SettingsConfigDict
from typing import Optional


class Settings(BaseSettings):
    """Type-safe configuration using Pydantic Settings."""
    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False
    )
    
    # Gemini Configuration
    gemini_api_key: str
    gemini_model: str = "gemini-2.5-flash"
    
    # Go Backend Configuration
    backend_url: str = "http://localhost:8080"
    
    # Feature Flags
    use_pinecone: bool = False
    
    # API Configuration
    api_host: str = "0.0.0.0"
    api_port: int = 8000
    cors_origins: list[str] = ["http://localhost:5173", "http://localhost:3000"]


settings = Settings()

