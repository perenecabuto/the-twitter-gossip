var React = require('react');
var ReactDOM = require('react-dom');
var mqtt = require('mqtt');
var AreaChart = require('react-d3-basic').AreaChart;
var LineChart = require('react-d3-basic').LineChart;
var ajar = require('ajar');

var serviceURL = window.location.hostname + ":8000";

var MessageManager = (function() {
    var webSocket;
    var listeners = [];

    function connectWebsocket() {
        console.log("connect webSocket");
        webSocket = new WebSocket("ws://" + serviceURL + "/events");
        webSocket.onopen = function(event) {
            console.log("open", event);
        };

        webSocket.onclose = function() {
            console.log("connection closed");
            setTimeout(connectWebsocket, 1000);
        }

        webSocket.onmessage = function(event) {
            console.log("onmessage");
            var data = JSON.parse(event.data);
            for (var i in listeners) {
                listeners[i](data);
            }
        };
    }

    connectWebsocket();

    return {
        onMessage: function(callback) {
            listeners.push(callback);
        }
    };
})();

var AddGossip = React.createClass({
    getInitialState: function() {
        return {
            label: this.props.label || "",
            subjects: "",
            classifiers: ""
        }
    },
    getClassifiersPayload: function() {
        var classifiers = {};
        var currentLabel = "";
        this.state.classifiers.split("\n").map(function(line) {
            line = line.trim();
            if (line[0] == ':') {
                currentLabel = line.substring(1);
                classifiers[currentLabel] = [];
            } else if (classifiers[currentLabel]) {
                classifiers[currentLabel].push(line);
            }
        });

        return classifiers;
    },
    handleSubmit: function(e) {
        e.preventDefault();
        if (this.state.label.trim() == "") {
            alert("label is empty");
            return;
        }
        if (this.state.subjects.trim() == "") {
            alert("subjects is empty");
            return;
        }
        var payload = {
            label: this.state.label,
            subjects: this.state.subjects.split(","),
            classifiers: this.getClassifiersPayload()
        };
        ajar.post(location.protocol + "//" + serviceURL + "/gossip/", payload).then(console.log);
    },
    render: function() {
        return (
        <fieldset>
            <legend>{this.state.label}</legend>
            <form onSubmit={this.handleSubmit}>
                <div className="form-group">
                <label>Label</label><br />
                <input className="form-control" onChange={(e) => this.setState({'label': e.target.value}) } />
                </div>

                <div className="form-group">
                <label>Subjects (comma separated)</label><br />
                <input className="form-control" onChange={(e) => this.setState({'subjects': e.target.value}) } />
                </div>

                <div className="form-group">
                <label>Classifiers (<a>description</a>)</label><br />
                <textarea className="form-control" onChange={(e) => this.setState({'classifiers': e.target.value}) } />
                </div>

                <button type="submit" className="btn btn-default">Save</button>
            </form>
        </fieldset>
        );
    }
});

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
            if (message.gossip !== undefined && message.gossip !== this.props.name) {
                return;
            }

            var item = {key: 'key', index: new Date()};
            if (this.props.topValue) {
                item.top = this.props.topValue;
            }

            for (var key in message.events) {
                var alreadyInChart = false;
                for (var i in chartSeries) {
                    alreadyInChart = alreadyInChart || chartSeries[i].field == key;
                    item[key] = message.events[key];
                }
                if (!alreadyInChart) {
                   chartSeries.push({field: key, name: key, color: this.getRandomColor()});
               }
            }

            var data = this.state.data || [];
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
        <div className="panel-body" ref={(ref) => this._el = ref}>no data</div>
        );
    }
});


var App = React.createClass({
    getInitialState: function() {
        return {
            gossips: {}
        };
    },
    componentDidMount: function() {
        MessageManager.onMessage(function(message) {
            if (this.state.gossips[message.gossip] === undefined) {
                this.state.gossips[message.gossip] = true;
                ReactDOM.render(
                    <div>
                    {Object.keys(this.state.gossips).map(function(label) {
                        return (
                        <div className="pull-left col-xs-12 col-sm-8 col-md-6 col-lg-6">
                            <div className="panel panel-default">
                                <div className="panel-heading">Gossip: {label}</div>
                                <MultLineChartBox name={label} />
                            </div>
                        </div>)
                    })}
                    </div>,
                    this._el)
            }
        }.bind(this));

    },
    render: function() {
        return (
        <div className="container">
            <h1>Dashboard</h1>
            <AddGossip />
            <div className="row" ref={(ref) => this._el = ref}>
                <div className="col-xs-12 col-sm-12 col-md-12 col-lg-12">Loading...</div>
            </div>
        </div>
        );
    }
});


ReactDOM.render(<App />, document.getElementById('content'))
