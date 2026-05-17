# src/data/Dockerfile — Python 数据服务容器

FROM python:3.10-slim

WORKDIR /app

# 依赖安装
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# 应用代码
COPY app/ ./app/
COPY alembic/ ./alembic/
COPY tests/ ./tests/

EXPOSE 8081

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8081"]