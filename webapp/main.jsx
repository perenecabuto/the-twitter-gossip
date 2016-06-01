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
            //console.log("onmessage");
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

var GossipForm = React.createClass({
    getInitialState: function() {
        return {
            label: this.props.label || "",
            subjects: "",
            classifiers: ""
        }
    },
    componentDidMount: function() {
        if (this.props.gossip) {
            ajar.get(location.protocol + "//" + serviceURL + "/gossip/" + this.props.gossip)
            .then(function(gossip) {
                this.setState({
                    gossip: gossip.gossip,
                    subjects: gossip.subjects.join(", "),
                    classifiers: Object.keys(gossip.classifiers).map(function(label) {
                        var patterns = gossip.classifiers[label];
                        return ":" + label + "\n" + patterns.join("\n");
                    }).join("\n")
                });
            }.bind(this));
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
        if (this.state.gossip.trim() == "") {
            alert("gossip name is empty");
            return;
        }
        if (this.state.subjects.trim() == "") {
            alert("subjects is empty");
            return;
        }
        var payload = {
            gossip: this.state.gossip,
            subjects: this.state.subjects.split(",").map((s) => s.trim()),
            classifiers: this.getClassifiersPayload()
        };
        ajar.post(location.protocol + "//" + serviceURL + "/gossip/", payload).then(function(data) {
            alert("gossip saved successfully");
        });
    },
    render: function() {
        return (
        <form onSubmit={this.handleSubmit}>
            <div className="form-group">
            <label>Gossip</label><br />
            <input className="form-control" value={this.state.gossip}
                onChange={(e) => this.setState({'gossip': e.target.value}) } />
            </div>

            <div className="form-group">
            <label>Subjects (comma separated)</label><br />
            <input className="form-control" value={this.state.subjects}
                onChange={(e) => this.setState({'subjects': e.target.value}) } />
            </div>

            <div className="form-group">
            <label>Classifiers (<a>description</a>)</label><br />
            <textarea className="form-control" value={this.state.classifiers}
                onChange={(e) => this.setState({'classifiers': e.target.value}) } />
            </div>

            <button type="submit" className="btn btn-default">Save</button>
        </form>
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
        if (this.props.topValue) {
            this.state.chartSeries.push({field: "top", name: "", color: "transparent"});
        }

        if (this.props.gossip) {
            ajar.get(location.protocol + "//" + serviceURL + "/gossip/" + this.props.gossip + "/history")
            .then(function(data) {
                data.history.reverse();
                for (var i in data.history) {
                    this.addChartItem(data.history[i]);
                }

                this.renderChart.call(this);
            }.bind(this));
        }

        MessageManager.onMessage(function(message) {
            if (!this.isMounted() || message.gossip !== undefined && message.gossip !== this.props.gossip) {
                return;
            }

            this.addChartItem(message);
            this.renderChart.call(this);
        }.bind(this));
    },

    addChartItem: function(gossipEvent) {
        var chartSeries = this.state.chartSeries;
        var item = {index: new Date(gossipEvent.timestamp * 1000)};
        if (this.props.topValue) {
            item.top = this.props.topValue;
        }

        for (var key in gossipEvent.events) {
            var alreadyInChart = false;
            for (var i in chartSeries) {
                alreadyInChart = alreadyInChart || chartSeries[i].field == key;
            }
            if (!alreadyInChart) {
                chartSeries.push({field: key, name: key, color: this.getRandomColor()});
            }
            item[key] = gossipEvent.events[key];
        }

        //console.log(item);

        var data = this.state.data || [];
        if (data.length == this.state.maxItems) {
            data.shift();
        }
        data.push(item);
        this.state.data = data;
    },
    render: function() {
        return (
        <div className="panel-body" ref={(ref) => this._el = ref}>no data</div>
        );
    }
});

var GossipPanel = React.createClass({
    getInitialState: function() {
        return {
            edit: false
        };
    },
    toggleTemplate: function() {
        this.setState({edit: !this.state.edit});
    },
    stopWorker: function() {
        ajar.get(location.protocol + "//" + serviceURL + "/gossip/" + this.props.gossip + "/stop").then(function(data) {
            alert("Worker state " + data.state);
        }.bind(this));
    },
    startWorker: function() {
        ajar.get(location.protocol + "//" + serviceURL + "/gossip/" + this.props.gossip + "/start").then(function() {
            alert("Worker state " + data.state);
        }.bind(this));
    },
    render: function() {
        var template;
        if (this.state.edit) {
            template = <GossipForm gossip={this.props.gossip} />;
        } else {
            template = <MultLineChartBox gossip={this.props.gossip} />;
        }
        return (
        <div className="pull-left col-xs-12 col-sm-8 col-md-6 col-lg-6">
            <div className="panel panel-default">
                <div className="panel-heading">
                    Gossip: {this.props.gossip}
                    <button type="button" className="pull-right" onClick={this.startWorker}>Start</button>
                    <button type="button" className="pull-right" onClick={this.stopWorker}>Stop</button>
                    <button type="button" className="pull-right" onClick={this.toggleTemplate}>Edit</button>
                </div>
                <div className="panel-body">
                    {template}
                </div>
            </div>
        </div>
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
        ajar.get(location.protocol + "//" + serviceURL + "/gossip/")
            .then(function(data) {
                for (var i in data.gossips) {
                    var item = data.gossips[i];
                    if (this.state.gossips[item.gossip] === undefined) {
                        this.state.gossips[item.gossip] = true;
                    }
                }
                this.setState({gossips: this.state.gossips});
            }.bind(this));
        MessageManager.onMessage(function(message) {
            if (this.state.gossips[message.gossip] === undefined) {
                this.state.gossips[message.gossip] = true;
            }
            this.setState({gossips: this.state.gossips});
        }.bind(this));

    },
    render: function() {
        return (
        <div className="container">
            <h1>Dashboard</h1>
            <div className="row" ref={(ref) => this._el = ref}>
                {Object.keys(this.state.gossips).map(function(gossip) {
                    return (<GossipPanel gossip={gossip} />)
                })}
            </div>
        </div>
        );
    }
});


ReactDOM.render(<App />, document.getElementById('content'))
