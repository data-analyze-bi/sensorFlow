import json
from datetime import datetime
from uuid import uuid4

from superset.app import create_app

DASH_TITLE = "埋点数据总览"
DATASET_KEY = {"table_name": "event", "schema": "sensors"}


def metric_count(label="事件数"):
    return {
        "aggregate": None,
        "column": None,
        "expressionType": "SQL",
        "hasCustomLabel": True,
        "label": label,
        "optionName": "metric_count",
        "sqlExpression": "count()",
    }


def metric_sql(label, sql):
    return {
        "aggregate": None,
        "column": None,
        "expressionType": "SQL",
        "hasCustomLabel": True,
        "label": label,
        "optionName": "metric_" + label,
        "sqlExpression": sql,
    }


def base_params(viz_type):
    return {
        "datasource": "1__table",
        "viz_type": viz_type,
        "time_range": "No filter",
        "adhoc_filters": [],
        "row_limit": 10000,
    }


def chart_specs():
    specs = []

    p = base_params("big_number_total")
    p.update({"metric": metric_count("总事件数"), "subheader": "全部埋点事件", "y_axis_format": "SMART_NUMBER"})
    specs.append(("总事件数", "big_number_total", p))

    p = base_params("big_number_total")
    p.update({"metric": metric_sql("独立设备数", "uniqExact(distinct_id)"), "subheader": "distinct_id 去重", "y_axis_format": "SMART_NUMBER"})
    specs.append(("独立设备数", "big_number_total", p))

    p = base_params("big_number_total")
    p.update({"metric": metric_sql("事件类型数", "uniqExact(event)"), "subheader": "event 去重", "y_axis_format": "SMART_NUMBER"})
    specs.append(("事件类型数", "big_number_total", p))

    p = base_params("echarts_timeseries_bar")
    p.update({
        "x_axis": "time",
        "time_grain_sqla": "PT1M",
        "metrics": [metric_count("事件数")],
        "groupby": [],
        "order_desc": True,
        "show_legend": False,
        "truncate_metric": True,
        "y_axis_format": "SMART_NUMBER",
        "rich_tooltip": True,
        "show_value": False,
    })
    specs.append(("每分钟事件趋势", "echarts_timeseries_bar", p))

    p = base_params("table")
    p.update({
        "query_mode": "aggregate",
        "groupby": ["event"],
        "metrics": [metric_count("事件数")],
        "order_by_cols": [json.dumps(["事件数", False])],
        "row_limit": 25,
        "server_page_length": 25,
        "show_cell_bars": True,
    })
    specs.append(("Top 事件排行", "table", p))

    p = base_params("pie")
    p.update({
        "groupby": ["$os"],
        "metric": metric_count("事件数"),
        "row_limit": 10,
        "show_labels": True,
        "show_legend": True,
        "label_type": "key_percent",
        "number_format": "SMART_NUMBER",
    })
    specs.append(("操作系统分布", "pie", p))

    p = base_params("table")
    p.update({
        "query_mode": "aggregate",
        "groupby": ["lt_page_name"],
        "metrics": [metric_count("事件数"), metric_sql("平均停留ms", "avg(toFloat64OrZero(lt_duration_ms))")],
        "order_by_cols": [json.dumps(["事件数", False])],
        "row_limit": 20,
        "server_page_length": 20,
        "show_cell_bars": True,
    })
    specs.append(("页面分布与平均耗时", "table", p))

    p = base_params("pie")
    p.update({
        "groupby": ["$network_type"],
        "metric": metric_count("事件数"),
        "row_limit": 10,
        "show_labels": True,
        "show_legend": True,
        "label_type": "key_percent",
        "number_format": "SMART_NUMBER",
    })
    specs.append(("网络类型分布", "pie", p))

    return specs


def make_position_json(chart_ids):
    root_id = "ROOT_ID"
    grid_id = "GRID_ID"
    row1, row2, row3 = "ROW-1", "ROW-2", "ROW-3"
    position = {
        "DASHBOARD_VERSION_KEY": "v2",
        root_id: {"type": "ROOT", "id": root_id, "children": [grid_id]},
        grid_id: {"type": "GRID", "id": grid_id, "children": [row1, row2, row3]},
        row1: {"type": "ROW", "id": row1, "children": [], "meta": {"background": "BACKGROUND_TRANSPARENT"}},
        row2: {"type": "ROW", "id": row2, "children": [], "meta": {"background": "BACKGROUND_TRANSPARENT"}},
        row3: {"type": "ROW", "id": row3, "children": [], "meta": {"background": "BACKGROUND_TRANSPARENT"}},
    }
    widths = [4, 4, 4, 12, 6, 6, 6, 6]
    heights = [16, 16, 16, 50, 50, 50, 45, 45]
    rows = [row1, row1, row1, row2, row2, row2, row3, row3]
    for idx, cid in enumerate(chart_ids):
        chart_node = f"CHART-{cid}"
        position[chart_node] = {
            "type": "CHART",
            "id": chart_node,
            "children": [],
            "meta": {"chartId": cid, "height": heights[idx], "width": widths[idx]},
        }
        position[rows[idx]]["children"].append(chart_node)
    return json.dumps(position)


def main():
    app = create_app()
    with app.app_context():
        from superset import db
        from superset.connectors.sqla.models import SqlaTable
        from superset.models.slice import Slice
        from superset.models.dashboard import Dashboard

        dataset = db.session.query(SqlaTable).filter_by(**DATASET_KEY).one()
        datasource = f"{dataset.id}__{dataset.type}"

        dash = db.session.query(Dashboard).filter_by(dashboard_title=DASH_TITLE).one_or_none()
        if dash is None:
            dash = Dashboard(dashboard_title=DASH_TITLE, slug="sensorflow-events-overview", published=True)
            db.session.add(dash)
        else:
            dash.slices = []
            for slc in db.session.query(Slice).filter(Slice.slice_name.like("埋点-%")).all():
                db.session.delete(slc)
            db.session.flush()

        slices = []
        for name, viz_type, params in chart_specs():
            params["datasource"] = datasource
            slc = Slice(
                slice_name="埋点-" + name,
                datasource_id=dataset.id,
                datasource_type=dataset.type,
                datasource_name=dataset.datasource_name,
                viz_type=viz_type,
                params=json.dumps(params, ensure_ascii=False),
                query_context=None,
                uuid=uuid4(),
                description="自动生成：SensorFlow 埋点 ClickHouse 查询图表",
                created_on=datetime.utcnow(),
                changed_on=datetime.utcnow(),
            )
            db.session.add(slc)
            slices.append(slc)
        db.session.flush()

        dash.slices = slices
        dash.position_json = make_position_json([s.id for s in slices])
        dash.json_metadata = json.dumps({"label_colors": {}, "timed_refresh_immune_slices": [], "expanded_slices": {}}, ensure_ascii=False)
        dash.published = True
        db.session.commit()
        print("dashboard_id", dash.id)
        print("slug", dash.slug)
        for slc in slices:
            print("chart", slc.id, slc.slice_name, slc.viz_type)


if __name__ == "__main__":
    main()
