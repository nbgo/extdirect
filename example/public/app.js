Ext.application({

	name: 'Example',

	launch: function () {
		var me = this,
			store, basicInfo;

		Ext.direct.Manager.addProvider(DirectApi.REMOTE_API);
		Ext.tip.QuickTipManager.init();

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
					xtype: 'container',
					flex: 1,
					layout: 'hbox',
					items: [
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
						},
						basicInfo = Ext.widget('form',{
							flex: 1,
							title: 'Basic Information',
							border: false,
							bodyPadding: 10,
							// configs for BasicForm
							api: {
								// The server-side method to call for load() requests
								load: 'Db.getBasicInfo',
								// The server-side must mark the submit handler as a 'formHandler'
								submit: 'Db.updateBasicInfo'
							},
							// specify the order for the passed params
							paramOrder: ['uid', 'foo'],
							dockedItems: [{
								dock: 'bottom',
								xtype: 'toolbar',
								ui: 'footer',
								style: 'margin: 0 5px 5px 0;',
								items: ['->', {
									text: 'Submit',
									handler: function () {
										basicInfo.getForm().submit({
											params: {
												foo: 'bar',
												uid: 34
											}
										});
									}
								}]
							}],
							defaultType: 'textfield',
							defaults: {
								anchor: '100%'
							},
							items: [{
								fieldLabel: 'Name',
								name: 'name'
							}, {
								fieldLabel: 'Email',
								msgTarget: 'side',
								vtype: 'email',
								name: 'email'
							}, {
								fieldLabel: 'Company',
								name: 'company'
							}]
						})
					]
				}
			]
		});

		basicInfo.getForm().load({
			// pass 2 arguments to server side getBasicInfo method (len=2)
			params: {
				foo: 'bar',
				uid: 34
			}
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