# OVSDB protocol implementation

Partial implementation of Open vSwitch Database (OVSDB) management protocol as defined
in RFC 7047, and some extensions defined in
[ovsdb-server document](https://docs.openvswitch.org/en/latest/ref/ovsdb-server.7/).

It operates on raw TCP connection, and does not provide any higher level abstractions. It
utilizes [SON-RPC 1.0 partial implementation for wired connections](https://github.com/kazmanavt/jsonrpc) library for low level communication.

## Features

### Implemented methods
- `list_dbs`
- `get_schema`
- `transact`
- `cancel`
- `monitor`
- `monitor_cond`
- `monitor_cond_since`
- `monitor_cancel`
- `echo`

### Unimplemented methods
- `lock`
- `steal`
- `unlock`
- `monitor_cond_change`
-  `get_server_id`
-  `set_db_change_aware`
-  `convert`

### Implemented transactions operations
- `Insert`
- `Select`
- `Update`
- `Mutate`
- `Delete`
- `Wait`
- `Abort`
- `Commit`
- `Comment`

### Unimplemented transactions operations
- `Assert`

This library provide support to maintain partial image of running OVSDB in client space.
It provides DB generation based on received scheme. Db package implements methods to apply
OVSDB notifications containing _table-updates2_ received after `monitor_cond_since` or
`monitor_cond` request.

## Usage

    // Create connection to OVSDB server
	jconn, err := jsonrpc.NewConnection("unix", "/var/run/openvswitch/db.sock", zapSugaredLogger)
	if err != nil {
		log.Errorw("fail to connect to ovsdb", zap.Error(err))
		return nil, err
	}
	ovsdbClient, err := ovsdb.NewClient(ovsdb.ClientConf{
		Log:  zapSugaredLogger,
		Conn: jconn})
	if err != nil {
		return nil, err
	}

    // Get list of available databases
    dbs := ovsdbClient.ListDbs()

    // Get OVSDB schema
    schema, err := ovsdbClient.GetSchema(dbs[0])
    if err != nil {
        return nil, err
    }

    // Create database
    db := db.NewDB(schema)

__(To be continued...)__