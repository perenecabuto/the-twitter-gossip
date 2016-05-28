var React = require('react');
var ReactDOM = require('react-dom');
var mqtt = require('mqtt');
var LineChart = require('react-d3-basic').LineChart;
var AreaChart = require('react-d3-basic').AreaChart;

var MessageManager = {
    client: mqtt.connect('mqtt://broker.mqttdashboard.com:8000'),
    subscriptions: {},
    subscribe: function(subscriptionTopic, callback) {
        if (!this.subscriptions[subscriptionTopic]) {
            console.log("subscribe: " + subscriptionTopic);
            this.client.subscribe(subscriptionTopic);
            this.subscriptions[subscriptionTopic] = true;
        }
        this.client.on('message', function(topic, message) {
            if (subscriptionTopic == topic) {
                callback(message.toString());
            }
        });
    }
};


var LineChartBox = React.createClass({
    getInitialState: function() {
        var maxEvents = 30;
        var data = [];
        for (var i = maxEvents; i > 0; i--) {
            var date = new Date(new Date().setSeconds(i * 5 * -1));
            data.push({key: this.props.name, value: 0, index: date, top: this.props.maxValue});
        }
        return {
            className: "panel panel-warning",
            maxEvents: maxEvents,
            data: data
        };
    },
    chartSeries: function() {
        return [
            {
                field: "value",
                color: "rgba(76, 175, 80, 0.9)",
                name: this.props.name
            },
            {
                field: "top",
                color: "transparent",
                name: "Top value: " + this.props.maxValue
            }
        ];
    },
    renderChart: function() {
        setTimeout(function() {
            ReactDOM.render(
                <AreaChart
                    width={520}
                    height={200}
                    xScale={"time"}
                    data={this.state.data}
                    chartSeries={this.chartSeries()}
                    x={(d) => d.index}
                />, this._el);
        }.bind(this));
    },
    componentDidMount: function() {
        MessageManager.subscribe(this.props.topic, function(message) {
            var data = this.state.data || [];
            data.shift();
            data.push({
                top: this.props.maxValue,
                key: this.props.name,
                value: parseInt(message),
                index: new Date()
            });

            this.setState({data: data, className: "panel panel-default"});
            this.renderChart.call(this);
        }.bind(this));
    },
    render: function() {
        var lastEvent = this.state.data[this.state.data.length - 1];
        return (
        <div className={this.state.className}>
            <div className="panel-heading">{this.props.name} LineChart</div>
            <div className="panel-body" ref={(ref) => this._el = ref} style={{marginLeft: "-10%"}} />
            <div className="panel-footer">
                Last value <b>{lastEvent.value}</b> at <b>{lastEvent.index.toString()}</b>
            </div>
        </div>
        );
    }
});

var MessageBox = React.createClass({
    getInitialState: function() {
        return {
            className: "panel panel-warning",
            message: "--"
        };
    },
    componentDidMount: function() {
        MessageManager.subscribe(this.props.topic, function(message) {
            this.setState({message: message, className: "panel panel-default"});
        }.bind(this));
    },
    render: function() {
        return (
        <div className={this.state.className}>
            <div className="panel-heading">{this.props.name} Message</div>
            <div className="panel-body">{this.state.message}</div>
        </div>
        );
    }
});

var App = React.createClass({
    render: function() {
        return (
        <div className="container">
            <h1 style={{paddingLeft: "12px"}}>Dashboard</h1>
            <div className="pull-left col-xs-12 col-sm-4 col-md-4 col-lg-2">
                <MessageBox name="Text" topic="mydome/text/value" />
                <MessageBox name="Humidity" topic="mydome/humidity/value" />
                <MessageBox name="Gas" topic="mydome/gas/value" />
            </div>

            <div className="pull-left col-xs-12 col-sm-8 col-md-8 col-lg-5">
                <LineChartBox name="Humidity" maxValue="100" topic="mydome/humidity/value" />
            </div>
            <div className="pull-left col-xs-12 col-sm-8 col-md-8 col-lg-5">
                <LineChartBox name="Temp" maxValue="60" topic="mydome/temp/value" />
            </div>
        </div>
        );
    }
});


ReactDOM.render(<App />, document.getElementById('content'))
