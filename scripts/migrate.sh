#!/bin/bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
CLICKHOUSE_HOST="127.0.0.1"
CLICKHOUSE_PORT="9000"
CLICKHOUSE_USER="default"
CLICKHOUSE_PASSWORD=""
CLICKHOUSE_DB="sensors"

# 打印带颜色的信息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 执行ClickHouse查询
clickhouse_query() {
    local query="$1"
    clickhouse-client --host="$CLICKHOUSE_HOST" \
                     --port="$CLICKHOUSE_PORT" \
                     --user="$CLICKHOUSE_USER" \
                     --password="$CLICKHOUSE_PASSWORD" \
                     --query="$query" 2>&1
}

# 检查ClickHouse连接
check_clickhouse() {
    print_info "检查ClickHouse连接..."
    result=$(clickhouse_query "SELECT 1")
    if [ "$result" = "1" ]; then
        print_success "ClickHouse连接正常"
        return 0
    else
        print_error "ClickHouse连接失败: $result"
        return 1
    fi
}

# 检查表是否存在
check_table_exists() {
    local table_name="$1"
    result=$(clickhouse_query "EXISTS TABLE ${CLICKHOUSE_DB}.${table_name}")
    echo "$result"
}

# 备份表
backup_table() {
    local table_name="$1"
    local backup_name="${table_name}_backup_$(date +%Y%m%d_%H%M%S)"

    print_info "备份表 ${table_name} -> ${backup_name}..."

    # 创建备份表
    clickhouse_query "CREATE TABLE ${CLICKHOUSE_DB}.${backup_name} AS ${CLICKHOUSE_DB}.${table_name}"
    if [ $? -ne 0 ]; then
        print_error "创建备份表失败"
        return 1
    fi

    # 复制数据
    clickhouse_query "INSERT INTO ${CLICKHOUSE_DB}.${backup_name} SELECT * FROM ${CLICKHOUSE_DB}.${table_name}"
    if [ $? -ne 0 ]; then
        print_error "复制数据失败"
        return 1
    fi

    # 检查记录数
    original_count=$(clickhouse_query "SELECT count() FROM ${CLICKHOUSE_DB}.${table_name}")
    backup_count=$(clickhouse_query "SELECT count() FROM ${CLICKHOUSE_DB}.${backup_name}")

    if [ "$original_count" = "$backup_count" ]; then
        print_success "备份完成: ${backup_name} (${backup_count} 条记录)"
        return 0
    else
        print_error "备份数据不一致: 原表=${original_count}, 备份=${backup_count}"
        return 1
    fi
}

# 创建新表
create_new_table() {
    local table_name="$1"
    local new_table_name="${table_name}_new"

    print_info "创建新表 ${new_table_name}..."

    if [ "$table_name" = "event" ]; then
        clickhouse_query "
        CREATE TABLE IF NOT EXISTS ${CLICKHOUSE_DB}.${new_table_name} (
            \`time\` DateTime DEFAULT now(),
            \`ds\` Date DEFAULT today(),
            \`event\` String,
            \`distinct_id\` String,
            \`\$os\` String DEFAULT '',
            \`\$os_version\` String DEFAULT '',
            \`\$model\` String DEFAULT '',
            \`\$brand\` String DEFAULT '',
            \`\$screen_width\` Nullable(Int32),
            \`\$screen_height\` Nullable(Int32),
            \`\$wifi\` Nullable(Bool),
            \`\$network_type\` String DEFAULT '',
            \`\$app_version\` String DEFAULT '',
            \`\$app_name\` String DEFAULT '',
            \`\$lib\` String DEFAULT '',
            \`\$lib_version\` String DEFAULT '',
            \`\$user_id\` String DEFAULT '',
            \`\$is_first_day\` Nullable(Bool),
            INDEX idx_event (\`event\`) TYPE bloom_filter GRANULARITY 1,
            INDEX idx_distinct_id (\`distinct_id\`) TYPE bloom_filter GRANULARITY 1,
            INDEX idx_time (\`time\`) TYPE minmax GRANULARITY 1,
            INDEX idx_ds (\`ds\`) TYPE minmax GRANULARITY 1
        ) ENGINE = MergeTree()
        PARTITION BY toYYYYMM(\`time\`)
        ORDER BY (\`distinct_id\`, \`event\`, \`time\`)
        TTL \`time\` + INTERVAL 1 YEAR
        SETTINGS index_granularity = 8192
        "
    elif [ "$table_name" = "user" ]; then
        clickhouse_query "
        CREATE TABLE IF NOT EXISTS ${CLICKHOUSE_DB}.${new_table_name} (
            \`distinct_id\` String,
            \`ds\` Date DEFAULT today(),
            \`user_id\` String DEFAULT '',
            \`first_visit_time\` Nullable(DateTime),
            \`last_visit_time\` DateTime DEFAULT now(),
            \`\$name\` String DEFAULT '',
            \`\$email\` String DEFAULT '',
            \`\$phone\` String DEFAULT '',
            \`\$avatar\` String DEFAULT '',
            INDEX idx_distinct_id (\`distinct_id\`) TYPE bloom_filter GRANULARITY 1,
            INDEX idx_user_id (\`user_id\`) TYPE bloom_filter GRANULARITY 1,
            INDEX idx_ds (\`ds\`) TYPE minmax GRANULARITY 1
        ) ENGINE = ReplacingMergeTree(last_visit_time)
        ORDER BY (\`distinct_id\`)
        SETTINGS index_granularity = 8192
        "
    fi

    if [ $? -eq 0 ]; then
        print_success "新表创建成功: ${new_table_name}"
        return 0
    else
        print_error "新表创建失败"
        return 1
    fi
}

# 迁移数据
migrate_data() {
    local table_name="$1"
    local old_table="${CLICKHOUSE_DB}.${table_name}"
    local new_table="${CLICKHOUSE_DB}.${table_name}_new"

    print_info "开始迁移数据: ${old_table} -> ${new_table}..."

    if [ "$table_name" = "event" ]; then
        # 获取动态列
        dynamic_columns=$(clickhouse_query "
        SELECT name FROM system.columns
        WHERE database = '${CLICKHOUSE_DB}' AND table = '${table_name}'
        AND name NOT IN (
            'time', 'ds', 'event', 'distinct_id',
            '\$os', '\$os_version', '\$model', '\$brand',
            '\$screen_width', '\$screen_height', '\$wifi', '\$network_type',
            '\$app_version', '\$app_name', '\$lib', '\$lib_version',
            '\$user_id', '\$is_first_day'
        )
        " | tr '\n' ',' | sed 's/,$//')

        # 构建INSERT语句
        clickhouse_query "
        INSERT INTO ${new_table}
        SELECT
            time,
            toDate(ds) as ds,
            event,
            distinct_id,
            \`\$os\`,
            \`\$os_version\`,
            \`\$model\`,
            \`\$brand\`,
            toInt32OrNull(\`\$screen_width\`) as \`\$screen_width\`,
            toInt32OrNull(\`\$screen_height\`) as \`\$screen_height\`,
            if(\`\$wifi\` = 'true', true, if(\`\$wifi\` = 'false', false, NULL)) as \`\$wifi\`,
            \`\$network_type\`,
            \`\$app_version\`,
            \`\$app_name\`,
            \`\$lib\`,
            \`\$lib_version\`,
            \`\$user_id\`,
            if(\`\$is_first_day\` = 'true', true, if(\`\$is_first_day\` = 'false', false, NULL)) as \`\$is_first_day\`
        FROM ${old_table}
        "
    elif [ "$table_name" = "user" ]; then
        clickhouse_query "
        INSERT INTO ${new_table}
        SELECT
            distinct_id,
            toDate(ds) as ds,
            user_id,
            first_visit_time,
            last_visit_time,
            \`\$name\`,
            \`\$email\`,
            \`\$phone\`,
            \`\$avatar\`
        FROM ${old_table}
        "
    fi

    if [ $? -eq 0 ]; then
        print_success "数据迁移完成"
        return 0
    else
        print_error "数据迁移失败"
        return 1
    fi
}

# 验证数据
verify_data() {
    local table_name="$1"
    local old_table="${CLICKHOUSE_DB}.${table_name}"
    local new_table="${CLICKHOUSE_DB}.${table_name}_new"

    print_info "验证数据一致性..."

    # 检查记录数
    old_count=$(clickhouse_query "SELECT count() FROM ${old_table}")
    new_count=$(clickhouse_query "SELECT count() FROM ${new_table}")

    print_info "原表记录数: ${old_count}"
    print_info "新表记录数: ${new_count}"

    if [ "$old_count" != "$new_count" ]; then
        print_error "记录数不一致！"
        return 1
    fi

    # 检查ds字段类型
    ds_type=$(clickhouse_query "SELECT toTypeName(ds) FROM ${new_table} LIMIT 1")
    print_info "ds字段类型: ${ds_type}"

    if [ "$ds_type" = "Date" ]; then
        print_success "ds字段类型正确"
    else
        print_error "ds字段类型错误: ${ds_type}"
        return 1
    fi

    print_success "数据验证通过"
    return 0
}

# 切换表
switch_tables() {
    local table_name="$1"
    local old_table="${table_name}"
    local new_table="${table_name}_new"
    local backup_table="${table_name}_old"

    print_warning "准备切换表..."
    read -p "是否确认切换？这将影响线上服务！(yes/no): " confirm

    if [ "$confirm" != "yes" ]; then
        print_info "取消切换"
        return 1
    fi

    print_info "切换表: ${old_table} -> ${backup_table}, ${new_table} -> ${old_table}"

    clickhouse_query "RENAME TABLE
        ${CLICKHOUSE_DB}.${old_table} TO ${CLICKHOUSE_DB}.${backup_table},
        ${CLICKHOUSE_DB}.${new_table} TO ${CLICKHOUSE_DB}.${old_table}
    "

    if [ $? -eq 0 ]; then
        print_success "表切换成功"
        return 0
    else
        print_error "表切换失败"
        return 1
    fi
}

# 主流程
main() {
    echo "=========================================="
    echo "  Sensors ClickHouse 表优化迁移脚本"
    echo "=========================================="
    echo ""

    # 检查连接
    check_clickhouse || exit 1

    # 选择操作
    echo ""
    echo "请选择操作："
    echo "1) 完整迁移流程（备份->创建新表->迁移数据->验证->切换）"
    echo "2) 仅备份表"
    echo "3) 创建新表"
    echo "4) 迁移数据"
    echo "5) 验证数据"
    echo "6) 切换表"
    echo "7) 退出"
    echo ""
    read -p "请输入选项 (1-7): " choice

    case $choice in
        1)
            # 完整流程
            for table in "event" "user"; do
                echo ""
                print_info "处理表: ${table}"

                # 备份
                backup_table "$table" || continue

                # 创建新表
                create_new_table "$table" || continue

                # 迁移数据
                migrate_data "$table" || continue

                # 验证
                verify_data "$table" || continue

                # 切换（需要确认）
                switch_tables "$table"
            done
            ;;
        2)
            read -p "请输入表名 (event/user): " table
            backup_table "$table"
            ;;
        3)
            read -p "请输入表名 (event/user): " table
            create_new_table "$table"
            ;;
        4)
            read -p "请输入表名 (event/user): " table
            migrate_data "$table"
            ;;
        5)
            read -p "请输入表名 (event/user): " table
            verify_data "$table"
            ;;
        6)
            read -p "请输入表名 (event/user): " table
            switch_tables "$table"
            ;;
        7)
            print_info "退出"
            exit 0
            ;;
        *)
            print_error "无效的选项"
            exit 1
            ;;
    esac

    echo ""
    print_success "操作完成！"
}

# 运行主流程
main
