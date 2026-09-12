import os

BABEL_DEFAULT_LOCALE = "zh"
BABEL_DEFAULT_FOLDER = "superset/translations"
ENABLE_PROXY_FIX = True

SUPERSET_WEBSERVER_PROTOCOL = "https"
SUPERSET_WEBSERVER_BASEURL = os.environ.get("SUPERSET_PUBLIC_BASE_URL", "http://localhost:8088/")

LANGUAGES = {
    "zh": {"flag": "cn", "name": "简体中文"},
    "en": {"flag": "us", "name": "English"},
}
