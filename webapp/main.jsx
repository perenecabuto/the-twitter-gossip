var React = require('react');
var ReactDOM = require('react-dom');
var mqtt = require('mqtt');
var AreaChart = require('react-d3-basic').AreaChart;
var LineChart = require('react-d3-basic').LineChart;


var MessageManager = (function() {
    var webSocket = new WebSocket("ws://localhost:8000/events");
    var listeners = [];

    webSocket.onopen = function(event) {
        console.log("open", event);
    };

    webSocket.onmessage = function(event) {
        var data = JSON.parse(event.data);
        for (var i in listeners) {
            listeners[i](data);
        }
    };

    return {
        onMessage: function(callback) {
            listeners.push(callback);
        }
    };
})();


var MultLineChartBox = React.createClass({
    getInitialState: function() {
        return {
            maxItems: 20,
            chartSeries: [],
            data: []
        };
    },
    getRandomColor: function() {
        var letters = '0123456789ABCDEF'.split('');
        var color = '#';
        for (var i = 0; i < 6; i++ ) {
            color += letters[Math.floor(Math.random() * 16)];
        }
        return color;
    },
    renderChart: function() {
        setTimeout(function() {
            ReactDOM.render(
                <div style={{marginLeft: "-10%"}}>
                    <LineChart
                        width={520}
                        height={200}
                        xScale={"time"}
                        data={this.state.data}
                        chartSeries={this.state.chartSeries}
                        x={(d) => d.index}
                    />
                </div>, this._el);
        }.bind(this));
    },
    componentDidMount: function() {
        var chartSeries = this.state.chartSeries;
        if (this.props.topValue) {
            chartSeries.push({field: "top", name: "", color: "transparent"});
        }
        MessageManager.onMessage(function(message) {
            var item = {key: 'key', index: new Date()};
            if (this.props.topValue) {
                item.top = this.props.topValue;
            }

            for (var key in message) {
                var alreadyInChart = false;
                for (var i in chartSeries) {
                    alreadyInChart = alreadyInChart || chartSeries[i].field == key;
                    item[key] = message[key];
                }
                if (!alreadyInChart) {
                   chartSeries.push({field: key, name: key, color: this.getRandomColor()});
               }
            }

            var data = this.state.data || [];
            console.log(data.length, this.state.maxItems);
            if (data.length == this.state.maxItems) {
                data.shift();
            }
            data.push(item);

            this.setState({data: data, chartSeries: chartSeries});
            this.renderChart.call(this);
        }.bind(this));
    },
    render: function() {
        return (
        <div className="panel panel-default">
            <div className="panel-heading">Realtime Chart</div>
            <div className="panel-body" ref={(ref) => this._el = ref}>no data</div>
            <div className="panel-footer">
            </div>
        </div>
        );
    }
});


var App = React.createClass({
    render: function() {
        return (
        <div className="container">
            <h1 style={{paddingLeft: "12px"}}>Dashboard</h1>
            <div className="pull-left col-xs-12 col-sm-6 col-md-6 col-lg-6">
                <MultLineChartBox name="gossip" topic="mydome/humidity/value" />
            </div>
        </div>
        );
    }
});


ReactDOM.render(<App />, document.getElementById('content'))
