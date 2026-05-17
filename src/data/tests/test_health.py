from fastapi.testclient import TestClient
import sys, os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))
from app.main import app
client = TestClient(app)

def test_health_returns_200():
    response = client.get("/data/v1/health")
    assert response.status_code == 200

def test_health_returns_code_0():
    response = client.get("/data/v1/health")
    body = response.json()
    assert body["code"] == 0
    assert body["data"]["status"] == "ok"

def test_root_returns_info():
    response = client.get("/")
    assert response.status_code == 200
