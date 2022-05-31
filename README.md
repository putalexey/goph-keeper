# Менеджер паролей GophKeeper

GophKeeper - выпускной проект Яндекс Практикума. Представляет собой клиент-серверную систему, позволяющую пользователю надёжно и безопасно хранить логины, пароли, бинарные данные и прочую приватную информацию.

## Сборка
В Makefile описаны следующие команды:
- make build_client - скомпилировать клиент
- make build_server - скомпилировать сервер
- make build - скомпилировать и клиент и сервер
- make proto - скомпилировать protobuf файлы
- make migrate - запустить миграции, сами миграции лежат в папке migrations

## GophKeeper server <small>(gk-server)</small>
Сервер принимает запросы по протоколу gRPC и сохраняет данные в базу PostgreSQL.
В репозитории лежит docker-compose.yml для запуска сервера. Достаточно создать рядом
с ним файл .env, в котором указать пароль для БД и запустить `docker-compose up -d`

```
> .env
POSTGRES_PASSWORD=secret
```

### Настройка
Сервер можно настраивать через переменные окружения, json конфиг и аргументы командной строки.<br>
Приоритет параметров `env < json < командная строка`

| description                                | env             | json            | Аргумент КС | default                          |
|--------------------------------------------|-----------------|-----------------|:-----------:|----------------------------------|
| Файл конфига в формате JSON                | CONFIG          |                 |     -c      | gk-server.json (в рабочей папке) |
| Адрес и порт, на котором принимать запросы | SERVER_ADDRESS  | server_address  |     -a      | :3030                            |
| Путь к файлу логов                         | LOGFILE_PATH    | logfile_path    |     -l      | вывод в консоль                  |
| Строка соединения с базой данных           | DATABASE_URI    | database        |     -d      |                                  |

> Пример строки соединения с базой данных: `postgres://db_user:db_pass@db_host:5432/db_name`

## GophKeeper client <small>(gk-client)</small>
### Настройка
Клиент также можно настраивать через переменные окружения, json конфиг и аргументы командной строки.<br>
Приоритет параметров `env < json < командная строка`

| description                 | env          | json         | Аргумент КС | default                           |
|-----------------------------|--------------|--------------|:-----------:|-----------------------------------|
| Файл конфига в формате JSON | CONFIG       |              |     -c      | gk-client.json (в рабочей папке)  |
| Адрес сервера               | SERVER_HOST  | server_host  |     -s      | goph-keeper.putalexey.ru:3030     |
| Путь к файлу данных         | STORAGE_PATH | storage_path |     -t      | ~/.config/gk-client/store.json    |
| Путь к файлу лога           | LOGFILE_PATH | logfile_path |     -l      | ~/.config/gk-client/gk-client.log |

> `~` - домашняя папка пользователя

```
Usage: gk-client <command> [command_arguments...]

Available commands:
    ping - tests connection to the server
    register - register new user and authorize
    add - add new record
    edit - edit field of the saved record
    delete - delete record
    auth - authorize user
    get - register new user and authorize
    list - list records
    help - show current help
```

### Команды клиента
#### *ping*
Test connection to the server

    gk-client ping

#### *register*
Register new user on the server. After successful registration client is authorized automatically

    gk-client register [username]

#### *auth*
Authorizes user on the server

    gk-client auth [login]

#### *get*
Show records. If type of record was file, is will be saved to provided path

    gk-client get [record_name]
    gk-client get <file_record_name> [filepath]

#### *add*
Saves new record to the server

    gk-client add [type]
    gk-client add text [record_name] [text] [comment]
    gk-client add file [record_name] [filepath] [comment]
    gk-client add login [record_name] [login] [password] [comment]
    gk-client add card [record_name]

#### *edit*
Edit field of the record saved earlier

    gk-client edit [record_name [field [value|filepath]]]

#### *delete*
Delete record from server

    gk-client delete [record_name]
