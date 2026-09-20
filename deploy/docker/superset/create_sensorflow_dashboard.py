import json
from datetime import datetime
from uuid import uuid4

from superset.app import create_app

DASH_TITLE = "SensorFlow 产品与运营演示看板"
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
        "adhoc_filters": [
            {
                "clause": "WHERE",
                "comparator": "sensorflow-demo",
                "datasourceWarning": False,
                "expressionType": "SIMPLE",
                "filterOptionName": "filter_demo_app",
                "fromFormData": True,
                "operator": "==",
                "sqlExpression": None,
                "subject": "app_id",
            }
        ],
        "row_limit": 10000,
    }


def chart_specs():
    specs = []

    p = base_params("big_number_total")
    p.update({"metric": metric_count("总事件数"), "subheader": "Demo events", "y_axis_format": "SMART_NUMBER"})
    specs.append(("总事件数", "big_number_total", p))

    p = base_params("big_number_total")
    p.update({"metric": metric_sql("用户数", "uniqExact(distinct_id)"), "subheader": "Demo users", "y_axis_format": "SMART_NUMBER"})
    specs.append(("用户数", "big_number_total", p))

    p = base_params("big_number_total")
    p.update({"metric": metric_sql("今日活跃用户", "uniqExactIf(distinct_id, toDate(time) = today())"), "subheader": "DAU", "y_axis_format": "SMART_NUMBER"})
    specs.append(("今日活跃用户", "big_number_total", p))

    p = base_params("big_number_total")
    p.update({"metric": metric_sql("新用户", "uniqExactIf(distinct_id, is_first_day = 1)"), "subheader": "First-day users", "y_axis_format": "SMART_NUMBER"})
    specs.append(("新用户", "big_number_total", p))

    p = base_params("big_number_total")
    p.update({"metric": metric_sql("购买人数", "uniqExactIf(distinct_id, event = 'demo_purchase')"), "subheader": "Purchasers", "y_axis_format": "SMART_NUMBER"})
    specs.append(("购买人数", "big_number_total", p))

    p = base_params("big_number_total")
    p.update({"metric": metric_sql("Demo GMV", "sum(revenue)"), "subheader": "Demo revenue", "y_axis_format": ",.2f"})
    specs.append(("Demo GMV", "big_number_total", p))

    p = base_params("big_number_total")
    p.update({"metric": metric_sql("付费转化率", "uniqExactIf(distinct_id, event = 'demo_purchase') / greatest(uniqExact(distinct_id), 1)"), "subheader": "Purchasers / users", "y_axis_format": ".1%"})
    specs.append(("付费转化率", "big_number_total", p))

    p = base_params("echarts_timeseries_bar")
    p.update({
        "x_axis": "demo_hour",
        "metrics": [metric_count("事件数")],
        "groupby": [],
        "order_desc": True,
        "show_legend": False,
        "truncate_metric": True,
        "y_axis_format": "SMART_NUMBER",
        "rich_tooltip": True,
        "show_value": False,
    })
    specs.append(("小时事件趋势", "echarts_timeseries_bar", p))

    p = base_params("echarts_timeseries_line")
    p.update({"x_axis": "demo_day", "metrics": [metric_sql("活跃用户", "uniqExact(distinct_id)")], "groupby": [], "show_legend": False, "rich_tooltip": True, "y_axis_format": "SMART_NUMBER"})
    specs.append(("日活跃用户趋势", "echarts_timeseries_line", p))

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

    p = base_params("table")
    p.update({"query_mode": "aggregate", "groupby": ["event"], "metrics": [metric_sql("用户数", "uniqExact(distinct_id)")], "adhoc_filters": base_params("table")["adhoc_filters"] + [{"clause": "WHERE", "comparator": ["demo_app_open", "demo_product_view", "demo_add_to_cart", "demo_purchase"], "datasourceWarning": False, "expressionType": "SIMPLE", "filterOptionName": "filter_funnel", "fromFormData": True, "operator": "IN", "sqlExpression": None, "subject": "event"}], "row_limit": 10, "server_page_length": 10, "show_cell_bars": True})
    specs.append(("核心行为漏斗阶段", "table", p))

    p = base_params("pie")
    p.update({
        "groupby": ["os"],
        "metric": metric_count("事件数"),
        "row_limit": 10,
        "show_labels": True,
        "show_legend": True,
        "label_type": "key_percent",
        "number_format": "SMART_NUMBER",
    })
    specs.append(("操作系统分布", "pie", p))

    for column, title in [("channel", "获客渠道分布"), ("country", "国家地区分布"), ("page_name", "页面访问分布")]:
        p = base_params("pie")
        p.update({"groupby": [column], "metric": metric_count("事件数"), "row_limit": 10, "show_labels": True, "show_legend": True, "label_type": "key_percent", "number_format": "SMART_NUMBER"})
        specs.append((title, "pie", p))

    p = base_params("table")
    p.update({
        "query_mode": "aggregate",
        "groupby": ["app_version"],
        "metrics": [metric_count("事件数")],
        "order_by_cols": [json.dumps(["事件数", False])],
        "row_limit": 20,
        "server_page_length": 20,
        "show_cell_bars": True,
    })
    specs.append(("应用版本分布", "table", p))

    p = base_params("pie")
    p.update({
        "groupby": ["network_type"],
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
    row_ids = [f"ROW-{index}" for index in range(1, 8)]
    position = {
        "DASHBOARD_VERSION_KEY": "v2",
        root_id: {"type": "ROOT", "id": root_id, "children": [grid_id]},
        grid_id: {"type": "GRID", "id": grid_id, "children": row_ids},
    }
    for row_id in row_ids:
        position[row_id] = {"type": "ROW", "id": row_id, "children": [], "meta": {"background": "BACKGROUND_TRANSPARENT"}}
    for idx, cid in enumerate(chart_ids):
        chart_node = f"CHART-{cid}"
        position[chart_node] = {
            "type": "CHART",
            "id": chart_node,
            "children": [],
            "meta": {"chartId": cid, "height": 18 if idx < 7 else 45, "width": 3 if idx < 4 else (4 if idx < 7 else 6)},
        }
        row_index = 0 if idx < 4 else (1 if idx < 7 else min(2 + (idx - 7) // 2, len(row_ids) - 1))
        position[row_ids[row_index]]["children"].append(chart_node)
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

        dash = db.session.query(Dashboard).filter_by(slug="sensorflow-events-overview").one_or_none()
        if dash is None:
            dash = Dashboard(dashboard_title=DASH_TITLE, slug="sensorflow-events-overview", published=True)
            db.session.add(dash)
        else:
            dash.dashboard_title = DASH_TITLE
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
                description="自动生成：仅展示 app_id=sensorflow-demo 的演示数据，不代表真实业务指标",
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
