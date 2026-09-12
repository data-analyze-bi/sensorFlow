import json
import os

from superset.app import create_app


DATABASE_NAME = "SensorFlow ClickHouse"
DATASET_NAME = "event"
SCHEMA = "sensors"


def main() -> None:
    app = create_app()
    uri = os.environ.get("CLICKHOUSE_SQLALCHEMY_URI", "clickhousedb://default:@10.66.66.1:8123/sensors")

    with app.app_context():
        from superset import db
        from superset.connectors.sqla.models import SqlaTable
        from superset.models.core import Database

        database = db.session.query(Database).filter_by(database_name=DATABASE_NAME).one_or_none()
        if database is None:
            database = Database(database_name=DATABASE_NAME)
            db.session.add(database)

        database.set_sqlalchemy_uri(uri)
        database.expose_in_sqllab = True
        database.allow_run_async = False
        database.extra = json.dumps(
            {
                "metadata_params": {},
                "engine_params": {},
                "metadata_cache_timeout": {},
                "schemas_allowed_for_file_upload": [],
            }
        )
        db.session.commit()

        dataset = (
            db.session.query(SqlaTable)
            .filter_by(table_name=DATASET_NAME, schema=SCHEMA, database_id=database.id)
            .one_or_none()
        )
        if dataset is None:
            dataset = SqlaTable(table_name=DATASET_NAME, schema=SCHEMA, database=database)
            db.session.add(dataset)
        dataset.is_sqllab_view = False
        dataset.fetch_metadata()
        db.session.commit()


if __name__ == "__main__":
    main()
