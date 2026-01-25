-- 诊断账号分配问题的 SQL 脚本

-- 1. 查看 group_id=4 的基本信息
SELECT id, name, created_at
FROM groups
WHERE id = 4;

-- 2. 查看所有 Anthropic 平台的账号
SELECT id, name, platform, type, status, schedulable,
       created_at, updated_at, deleted_at
FROM accounts
WHERE platform = 'anthropic'
ORDER BY created_at DESC;

-- 3. 查看 group_id=4 关联的所有账号
SELECT ag.id, ag.group_id, ag.account_id, ag.priority,
       a.name as account_name, a.platform, a.status, a.schedulable
FROM account_groups ag
LEFT JOIN accounts a ON ag.account_id = a.id
WHERE ag.group_id = 4
ORDER BY ag.priority;

-- 4. 查看最近创建的账号(不在任何组中)
SELECT a.id, a.name, a.platform, a.type, a.status, a.schedulable,
       a.created_at,
       CASE
         WHEN EXISTS (SELECT 1 FROM account_groups WHERE account_id = a.id)
         THEN '已分配'
         ELSE '未分配'
       END as assignment_status
FROM accounts a
WHERE a.platform = 'anthropic'
  AND a.deleted_at IS NULL
ORDER BY a.created_at DESC
LIMIT 10;

-- 5. 检查账号的详细状态(可能影响调度的字段)
SELECT id, name, platform, status, schedulable,
       overload_until, rate_limit_reset_at,
       expired_at, temp_unschedulable_until,
       created_at
FROM accounts
WHERE platform = 'anthropic'
  AND deleted_at IS NULL
ORDER BY created_at DESC;
