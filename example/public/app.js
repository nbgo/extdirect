Ext.application({

	name: 'Example',

	launch: function () {
		var me = this,
			store;

		Ext.direct.Manager.addProvider(DirectApi.REMOTE_API);

		store = Ext.create('Ext.data.Store', {
			fields: ['id', 'text'],
			pageSize: 25,
			remoteSort: true,
			remoteFilter: true,
			autoLoad: false,
			proxy: {
				type: 'direct',
				api: {
					read: 'DirectApi.Db.getRecords'
				},
				reader: {
					type: 'json',
					rootProperty: 'records'
				},
				writer: {
					type: 'json',
					allowSingle: false,
					rootProperty: 'records'
				},
				extraParams: {
					model: 'User'
				}
			}
		});

		Ext.widget('viewport', {
			layout: {
				type: 'vbox',
				align: 'stretch'
			},
			items: [
				me.getTestRunnerCfg('test'),
				me.getTestRunnerCfg('testEcho1', 'Hello, Go!'),
				me.getTestRunnerCfg('testEcho2', 'Hello', 1, 2, 3, 4, 5, '!'),
				me.getTestRunnerCfg('testException1'),
				me.getTestRunnerCfg('testException2'),
				me.getTestRunnerCfg('testException3'),
				me.getTestRunnerCfg('testException4'),
				{
					xtype: 'button',
					text: 'RUN ALL',
					handler: function (btn) {
						btn.up('viewport').query('#testRunnerBtn').forEach(function (x) {
							x.handler(x);
						});
					}
				},
				{
					xtype: 'grid',
					flex: 1,
					store: store,
					columns: [
						{
							text: 'Id',
							dataIndex: 'id',
							width: 100
						},
						{
							text: 'Text',
							dataIndex: 'text',
							flex: 1
						}
					],
					tbar: [
						{
							xtype: 'button',
							text: 'Load',
							handler: function () {
								store.load();
							}
						},
						{
							xtype: 'button',
							text: 'Reload',
							handler: function () {
								store.reload();
							}
						}
					]
				}
			]
		});
	},

	getTestRunnerCfg: function (method) {
		var args = Array.prototype.slice.call(arguments, 1);
		return {
			xtype: 'container',
			layout: 'hbox',
			items: [
				{
					xtype: 'button',
					itemId: 'testRunnerBtn',
					text: method,
					width: 150,
					handler: function (btn) {
						args.push(function (data, response, success) {
							console.info(arguments);
							btn.up('container').down('displayfield').setValue(Example.app.stringifyDirectApiResponse(data, response, success));
						});
						DirectApi.Db[method].apply(DirectApi.Db, args);
					}
				},
				{
					xtype: 'tbspacer',
					width: 10
				},
				{
					xtype: 'displayfield',
					flex: 1
				}
			]
		};
	},

	stringifyDirectApiResponse: function (data, response, success) {
		return Ext.String.format('data = {0}; response = {1}; success = {2}', JSON.stringify(data), JSON.stringify(response), success);
	}
});