import asyncio
import threading
import uvicorn
from fastapi import FastAPI
from grpc import aio
from server.grpc_server import StrategyServicer
from strategy_pb import strategy_pb2_grpc

app = FastAPI(title="千川策略服务")


@app.get("/health")
def health():
    return {"status": "ok"}


async def serve_grpc():
    server = aio.server()
    strategy_pb2_grpc.add_StrategyServiceServicer_to_server(StrategyServicer(), server)
    server.add_insecure_port("[::]:50051")
    await server.start()
    await server.wait_for_termination()


def run_grpc():
    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)
    loop.run_until_complete(serve_grpc())


if __name__ == "__main__":
    grpc_thread = threading.Thread(target=run_grpc, daemon=True)
    grpc_thread.start()
    print("gRPC server starting on :50051")
    uvicorn.run(app, host="0.0.0.0", port=8000)
