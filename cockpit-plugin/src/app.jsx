/*
 * This file is part of Cockpit.
 *
 * Copyright (C) 2017 Red Hat, Inc.
 *
 * Cockpit is free software; you can redistribute it and/or modify it
 * under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation; either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Cockpit is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
 * Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Cockpit; If not, see <http://www.gnu.org/licenses/>.
 */

import React from 'react';
import { Card, CardBody, CardTitle, Tabs, Tab } from "@patternfly/react-core";
import { YamlFile } from './YamlFile.jsx';

// import cockpit from 'cockpit';
// const _ = cockpit.gettext;

export class Application extends React.Component {
    handleTabClick = (event, tabIndex) => {
        Tabs.setActiveTabKey(tabIndex);
    };

    render() {
        return (
            <Card>
                <CardTitle>MicroShift Configuration</CardTitle>
                <CardBody>
                    <Tabs defaultActiveKey={0} isFilled>
                        <Tab eventKey={0} title="MicroShift">
                            <YamlFile
                                fileName="/etc/microshift/config.yaml"
                            />
                        </Tab>
                        <Tab eventKey={1} title="LVMD">
                            <YamlFile
                                fileName="/etc/microshift/lvmd.yaml"
                            />
                        </Tab>
                        <Tab eventKey={2} title="OVN">
                            <YamlFile
                                fileName="/etc/microshift/ovn.yaml"
                            />
                        </Tab>
                    </Tabs>
                </CardBody>
            </Card>
        );
    }
}
