#!/usr/bin/env python3
# pylint: disable=too-few-public-methods, broad-except, line-too-long
"""
This script generates trend charts for bugs in MicroShift. It groups
created and resolved bugs per week over the last 6 months from the
moment the script is executed, producing a chart with results for both
metrics and the trend.

Args:
    To use this script, provide the following arguments:

    -t, --token, JIRA Auth token. Defaults to JIRA_TOKEN env var

File : bugtrend.py
"""

import jira
import os
import argparse
from datetime import date, datetime, timedelta
from dateutil.relativedelta import relativedelta
import matplotlib.pyplot as plt


JQL_FILTER_QUERY = 'filter = "MicroShift - Bugs in Project" order by createdDate asc'
JIRA_SERVER = 'https://issues.redhat.com'

class DayIssues:
    """
    Basic container to store created and resolved issues.
    Stores issue keys in case they are useful.
    """
    def __init__(self):
        self._created = []
        self._resolved = []
    def add_created(self, issue):
        self._created.append(issue)
    def add_resolved(self, issue):
        self._resolved.append(issue)
    def amount_created(self):
        return len(self._created)
    def amount_resolved(self):
        return len(self._resolved)

def parse_date_week(date):
    """
    Computes the week number from an ISO date format.
    """
    d = datetime.fromisoformat(date[:-5])
    return d.isocalendar()[1]

def daterange_week(start_date, end_date):
    """
    Generates week numbers between two given dates.
    """
    for n in range(int((end_date - start_date).days)//7):
        d = start_date+timedelta(n*7)
        yield d.isocalendar()[1]

def x_formatter(start, end):
    """
    Plot x axis formatter to display as dates between start and end dates.
    """
    def _internal(x, pos):
        int((end_date - start_date).days)//7
        d = start_date+timedelta(x*7)
        return d.strftime('%Y-%m-%d')
    return _internal

if __name__ == '__main__':
    parser = argparse.ArgumentParser(
        prog='bugtrend',
        description='MicroShift bug trend chart generator'
    )
    parser.add_argument(
        '-t',
        '--token',
        help='JIRA Auth token. Defaults to JIRA_TOKEN env var',
        default=os.environ['JIRA_TOKEN'])
    args = parser.parse_args()

    connection = jira.JIRA(
        server=JIRA_SERVER,
        token_auth=args.token)

    try:
        query = connection.search_issues(
            jql_str=JQL_FILTER_QUERY, maxResults=9999, expand='changelog')
        print(f"Scanning {len(query)} issues")
    except Exception as e:
        print(f"Unable to retrieve issues: {e}")
        exit(1)

    data = dict()
    total_created, total_resolved = 0, 0
    for issue in query:
        creation_date = parse_date_week(issue.fields.created)
        if creation_date not in data:
            data[creation_date] = DayIssues()
        data[creation_date].add_created(issue.key)
        total_created += 1

        for history in issue.changelog.histories:
            found = False
            for item in history.items:
                if (  item.field == 'status' \
                       and item.toString in ['Verified', 'Closed']) or \
                   (  item.field == 'resolution' \
                       and item.toString == 'Done'):
                    resolved_date = parse_date_week(history.created)
                    if resolved_date not in data:
                        data[resolved_date] = DayIssues()
                    data[resolved_date].add_resolved(issue.key)
                    total_resolved += 1
                    found = True
                    break
            if found:
                break

    end_date = date.today()
    start_date = end_date + relativedelta(days=-180)

    dates, created, resolved, trend = [], [], [], []
    for i, week in enumerate(daterange_week(start_date, end_date)):
        dates.append(i)
        i+=1
        if week in data:
            created.append(data[week].amount_created())
            resolved.append(data[week].amount_resolved())
        else:
            created.append(0)
            resolved.append(0)
        if len(trend) == 0:
            trend.append(created[-1]-resolved[-1])
        else:
            trend.append(trend[-1]+created[-1]-resolved[-1])

    fig, (ax1, ax2) = plt.subplots(2, 1, sharex=True)
    fig.suptitle(f'Issues created {total_created}. Issues resolved {total_resolved}. Since {start_date}')

    fmt = x_formatter(start_date, end_date)
    ax1.xaxis.set_major_formatter(fmt)
    ax2.xaxis.set_major_formatter(fmt)

    ax1.plot(dates, created, 'r')
    ax1.plot(dates, resolved, 'g')
    ax1.fill_between(
        dates,
        created,
        resolved,
        where=[resolved[i]>=created[i] for i in range(len(created))],
        facecolor='green',
        interpolate=True,
        alpha=0.5)
    ax1.fill_between(
        dates,
        created,
        resolved,
        where=[resolved[i]<=created[i] for i in range(len(created))],
        facecolor='red',
        interpolate=True,
        alpha=0.5)
    ax2.plot(dates, trend, 'b')
    ax2.yaxis.set_visible(False)

    fig.autofmt_xdate()
    plt.show()