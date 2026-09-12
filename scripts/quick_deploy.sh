#!/bin/bash

# 快速部署脚本 - 一键应用优化

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

print_step() {
    echo -e "${BLUE}[步骤 $1]${NC} $2"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# 检查环境
check_environment() {
    print_step 1 "检查环境"

    # 检查Go
    if ! command -v go &> /dev/null; then
        print_error "Go未安装"
        exit 1
    fi
    print_success "Go版本: $(go version)"

    # 检查ClickHouse客户端
    if ! command -v clickhouse-client &> /dev/null; then
        print_warning "ClickHouse客户端未安装，部分功能可能无法使用"
    else
        print_success "ClickHouse客户端已安装"
    fi

    echo ""
}

# 备份代码
backup_code() {
    print_step 2 "备份当前代码"

    if [ -d ".git" ]; then
        git add .
        git commit -m "backup before optimization" || true
        git tag "backup-$(date +%Y%m%d-%H%M%S)" || true
        print_success "代码已备份到Git"
    else
        tar -czf "../sensors-backup-$(date +%Y%m%d-%H%M%S).tar.gz" .
        print_success "代码已备份到tar.gz"
    fi

    echo ""
}

# 运行单元测试
run_tests() {
    print_step 3 "运行单元测试"

    if go test ./common/type_detector_test.go ./common/type_detector.go -v; then
        print_success "单元测试通过"
    else
        print_error "单元测试失败"
        exit 1
    fi

    echo ""
}

# 编译项目
build_project() {
    print_step 4 "编译项目"

    if go build -o sensors main.go; then
        print_success "编译成功"
    else
        print_error "编译失败"
        exit 1
    fi

    echo ""
}

# 显示下一步操作
show_next_steps() {
    print_step 5 "下一步操作"

    cat << EOF

${GREEN}代码优化已准备就绪！${NC}

接下来请选择部署方式：

${YELLOW}方式1: 新系统部署（推荐）${NC}
如果是全新部署，直接使用优化版表结构：
  1. 使用 sql/init_clickhouse.sql 初始化数据库
  2. 启动服务: ./sensors
  3. 发送测试数据验证

${YELLOW}方式2: 已有系统迁移${NC}
如果已有数据需要迁移：
  1. 运行迁移脚本: bash scripts/migrate.sh
  2. 选择 "完整迁移流程"
  3. 按提示完成数据迁移和验证
  4. 切换到新表
  5. 重启服务

${BLUE}性能测试：${NC}
完成部署后，运行性能测试：
  bash scripts/benchmark.sh

EOF
}

# 主函数
main() {
    echo "=========================================="
    echo "  Sensors 优化快速部署脚本"
    echo "=========================================="
    echo ""

    # 确认执行
    read -p "是否继续？(yes/no): " confirm
    if [ "$confirm" != "yes" ]; then
        echo "取消部署"
        exit 0
    fi

    echo ""

    # 执行步骤
    check_environment
    backup_code
    run_tests
    build_project
    show_next_steps

    echo ""
    print_success "部署准备完成！"
}

main
