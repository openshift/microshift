import cockpit from 'cockpit';
import React from 'react';
import { TextArea, Button, Alert, Split, SplitItem, Stack, StackItem } from "@patternfly/react-core";
const JSYaml = require("js-yaml");

export class YamlFile extends React.Component {
    constructor(props) {
        super(props);
        this.state = { content: "", modified: false, hasError: false, errMessage: "" };

        this.loadContent();
    }

    updateState = (c, m = false, he = false, em = "No error") => {
        this.setState({ content: c, modified: m, hasError: he, errMessage: em });
    };

    loadContent = () => {
        // Check cockpit.channel for permission elevation
        // https://cockpit-project.org/guide/latest/cockpit-channels.html#cockpit-channels-channel
        let fileName = this.props.fileName;

        const promise = cockpit.file(fileName).read();
        promise.then((content, tag) => {
            if (!content) {
                fileName += ".default";
                const promise = cockpit.file(fileName).read();
                promise.then((content, tag) => {
                    if (content) {
                        this.updateState(content);
                        this.validateContent();
                    } else {
                        this.updateState("", false, true, "Failed to open " + fileName);
                    }
                });
            } else {
                this.updateState(content);
                this.validateContent();
            }
        });
    };

    validateContent = () => {
        try {
            JSYaml.load(this.state.content);
            this.updateState(this.state.content, this.state.modified);
        } catch {
            this.updateState(this.state.content, this.state.modified, true, "Invalid YAML format");
        }
    };

    handleChangeText = (value) => {
        this.updateState(value, true, this.state.hasError, this.state.errMessage);
    };

    handleReload = (event) => {
        this.loadContent();
    };

    handleValidate = (event) => {
        this.validateContent();
    };

    handleSaveClick = (event) => {
        this.validateContent();
        // TODO: Elevate permissions and save if valid
    };

    render() {
        const taStyle = { height: "500px" };
        return (
            <Stack hasGutter>
                <StackItem>
                    <Alert
                        variant={this.state.hasError ? "danger" : "default"}
                        title={this.state.errMessage === "" ? "No errors" : this.state.errMessage}
                    />
                </StackItem>
                <StackItem>
                    <TextArea
                        id={this.props.fileName}
                        value={this.state.content}
                        onChange={value => this.handleChangeText(value)}
                        resizeOrientation='vertical'
                        style={taStyle}
                    />
                </StackItem>
                <StackItem>
                    {this.renderButtons()}
                </StackItem>
            </Stack>
        );
    }

    renderButtons = () => {
        return (
            <Split hasGutter>
                <SplitItem>
                    <Button
                        onClick={this.handleReload}
                    >
                        Reload
                    </Button>
                </SplitItem>
                <SplitItem>
                    <Button
                        isDisabled={this.state.content === ""}
                        onClick={this.handleValidate}
                    >
                        Validate
                    </Button>
                </SplitItem>
                <SplitItem>
                    <Button
                        isDisabled={!this.state.modified || this.state.hasError}
                        onClick={this.handleSaveClick}
                    >
                        Save
                    </Button>
                </SplitItem>
            </Split>
        );
    };
}
