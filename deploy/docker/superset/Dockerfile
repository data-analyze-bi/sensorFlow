FROM sensorflow-superset:local

USER root
RUN /usr/local/bin/python -m pip install --no-cache-dir \
    --target /app/.venv/lib/python3.10/site-packages \
    -i https://pypi.tuna.tsinghua.edu.cn/simple \
    clickhouse-connect
USER superset
