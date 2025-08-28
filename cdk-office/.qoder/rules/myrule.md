---
trigger: always_on
alwaysApply: true
---
{
"trigger": "always_on",
"alwaysApply": true,
"rules": [
{
"condition": "environment",
"type": "programming",
"value": "windows11_64bit"
},
{
"condition": "environment",
"type": "runtime",
"value": "windows11_64bit"
},
{
"condition": "script_syntax",
"type": "shell",
"value": "powershell"
},
{
"condition": "path",
"type": "project_directory",
"value": "D:\Downloads\Dify-Cdk-Office\cdk-office"
},
{
"condition": "path",
"type": "cdk_project",
"value": "D:\Downloads\Dify-Cdk-Office\cdk-master"
},
{
"condition": "path",
"type": "development_document",
"value": "D:\Downloads\Dify-Cdk-Office\.qoder\quests\dify-development-documentation.md"
},
{
"condition": "action",
"type": "compilation",
"value": "compile_after_module_completion"
},
{
"condition": "command",
"type": "path_usage",
"value": "always_use_absolute_paths"
},
{
"condition": "command",
"type": "multiple_commands",
"value": "use_semicolon_separator"
},
{
"condition": "code_modification",
"type": "logic",
"value": "use_complete_logic_no_simplification"
},
{
"condition": "git",
"type": "push_scope",
"value": "cdk-office-repo_directory_only"
}
]
}