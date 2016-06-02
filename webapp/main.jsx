var React = require('react');
var ReactDOM = require('react-dom');
var ajar = require('ajar');
var nv = require('nvd3');


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

            <button type="submit" className="btn btn-default pull-right">Save</button>
        </form>
        );
    }
});


var MultLineChartBox = React.createClass({
    getInitialState: function() {
        return {
            maxItems: 20,
            data: []
        };
    },
    getRandomColor: function() {
        this.colorCount = this.colorCount || 0;
        var colors = [
            '#3B7A57',
            '#00C4B0',
            '#FFBF00',
            '#FF7E00',
            '#FF033E',
            '#9966CC',
            '#A4C639'
        ]
        return colors[this.colorCount++];
    },
    renderChart: function() {
        var tickMultiFormat = d3.time.format.multi([
            ["%H:%M:%S", function(d) { return d.getMinutes() == 0; }],
            ["%M:%S", function(d) { return d.getSeconds() == 0; }],
            [":%S", function(d) { return true; }],
        ]);

        nv.addGraph(function() {
            var chart = nv.models.lineChart().options({duration: 300});
            nv.utils.windowResize(chart.update);

            chart.yAxis.axisLabel('Hits');
            chart.xAxis.axisLabel("Time")
            .tickFormat(function(d) {
                return tickMultiFormat(new Date(d));
            });

            d3.select(this._el).datum(this.state.data).call(chart);

            this.chart = chart;
            return chart;
        }.bind(this));
    },
    addFieldValue: function(field, value) {
        var data = this.state.data;
        var fieldData;
        for (var i in data) {
            if (data[i].key === field) {
                fieldData = data[i];
                break;
            }
        }

        if (fieldData === undefined) {
            fieldData = {key: field, values: [], color: this.getRandomColor()};
            data.push(fieldData);
            data.sort(function(a, b) {
                return a.key < b.key ? -1 : (a.key > b.key ? 1 : 0);
            });
        }

        if (fieldData.values.length >= this.state.maxItems) {
            fieldData.values.shift();
        }

        fieldData.values.push(value);
    },
    componentDidMount: function() {
        this.renderChart();
        if (this.props.topValue) {
            this.state.data.push({field: "top", key: "", color: "transparent", values: []});
        }

        MessageManager.onMessage(function(message) {
            if (!this.isMounted() || message.gossip !== undefined && message.gossip !== this.props.gossip) {
                return;
            }

            if (this.props.topValue) {
                this.addFieldValue("top", this.props.topValue);
            }

            for (var key in message.events) {
                this.addFieldValue(key, {x: new Date().getTime(), y: message.events[key]});
            }

            this.chart.update();
        }.bind(this));
    },
    render: function() {
        return (<svg style={{height: '300px', width: '100%'}} ref={(ref) => this._el = ref}></svg>);
    }
});

var GossipPanel = React.createClass({
    getInitialState: function() {
        return {
            edit: this.props.gossip === undefined
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
        ajar.get(location.protocol + "//" + serviceURL + "/gossip/").then(function(data) {
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
    addGossip: function() {
    },
    render: function() {
        return (
        <div className="container">
            <h1>Dashboard</h1>
            <button type="button" className="btn" onClick={this.addGossip}>New gossip</button>
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
