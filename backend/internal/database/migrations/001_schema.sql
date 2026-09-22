-- 帝王世系图谱数据库结构
-- 该脚本可重复执行（幂等）。

CREATE TABLE IF NOT EXISTS dynasty (
    id          BIGINT       NOT NULL AUTO_INCREMENT,
    name        VARCHAR(64)  NOT NULL,
    start_year  INT          NULL COMMENT '起始年，负数表示公元前',
    end_year    INT          NULL COMMENT '结束年，负数表示公元前',
    capital     VARCHAR(128) NOT NULL DEFAULT '',
    description TEXT         NULL,
    sort_order  INT          NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    UNIQUE KEY uk_dynasty_name (name),
    KEY idx_dynasty_sort (sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS emperor (
    id              BIGINT       NOT NULL AUTO_INCREMENT,
    dynasty_id      BIGINT       NOT NULL,
    name            VARCHAR(64)  NOT NULL COMMENT '姓名',
    temple_name     VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '庙号',
    posthumous_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '谥号',
    era_names       VARCHAR(255) NOT NULL DEFAULT '' COMMENT '年号',
    father_id       BIGINT       NULL COMMENT '父帝 id，自引用',
    order_index     INT          NOT NULL DEFAULT 0 COMMENT '在位顺序',
    relation_note   VARCHAR(128) NOT NULL DEFAULT '' COMMENT '继位关系说明',
    reign_start     INT          NULL COMMENT '在位起始年，负数为公元前',
    reign_end       INT          NULL COMMENT '在位结束年',
    birth_year      INT          NULL,
    death_year      INT          NULL,
    description     TEXT         NULL,
    PRIMARY KEY (id),
    KEY idx_emperor_dynasty (dynasty_id),
    KEY idx_emperor_father (father_id),
    KEY idx_emperor_order (dynasty_id, order_index),
    CONSTRAINT fk_emperor_dynasty FOREIGN KEY (dynasty_id) REFERENCES dynasty (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
