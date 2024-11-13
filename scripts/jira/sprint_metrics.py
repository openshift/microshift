#!/usr/bin/python3


"""Tool for extracing Sprint Metrics from JIRA.

It is mandatory to provide a configuration file in JSON format. The file needs to have this format:
{
    "board_id": board id from JIRA. This varies between projects,
    "jira_query": JQL to use for searching issues,
    "jira_token": Auth token from JIRA,
    "results_path": Path for the CSV results file,
    "raw_issues_path": Path for a raw JSON file with all sprints and issues that were retrieved.,
    "sprints": [
        {
            "name": Sprint name. Needs to match the name in JIRA,
            "goals_achieved": Number of goals that were achieved,
            "goals_total": Total number of goals that were included in the sprint,
            "team_members": Number of team members participating in the sprint's issues,
            "working_days": Sum of working days from every team member
        }
    ]
}

After running the script you may upload the CSV to a LLM model, using the prompts in AI_PROMPTS variable.
"""

import argparse
import csv
from jira import JIRA
import json
import logging

# TODO investigate whats wrong with maxResults from sprints in the api.
SKIP_FIRST_N_SPRINTS = 50

# Sprint name
CSV_FIELD_SPRINT_NAME = 'Sprint'
# Planned story points for a sprint
CSV_FIELD_PLANNED_TOTAL = 'Planned story points'
# Resolved story points at the end of a sprint
CSV_FIELD_RESOLVED_TOTAL = 'Velocity'
# Not started story points at the end of a sprint
CSV_FIELD_NOT_STARTED_TOTAL = 'Not started story points'
# In progress story points at the end of a sprint
CSV_FIELD_IN_PROGRESS_TOTAL = 'In progress story points'
# Waiting code review story points at the end of a sprint
CSV_FIELD_CODE_REVIEW_TOTAL = 'Waiting code review story points'
# Total bug story points at the end of a sprint
CSV_FIELD_BUG_TOTAL = 'Total bugs story points'
# Resolved bug story points at the end of a sprint
CSV_FIELD_BUG_RESOLVED = 'Bugs resolved story points'
# Bugs waiting verification story points at the end of a sprint
CSV_FIELD_BUG_WAITING_QE = 'Bugs waiting verification story points'
# Bugs waiting code review story points at the end of a sprint
CSV_FIELD_BUG_WAITING_CODE_REVIEW = 'Bugs waiting code review story points'
# Bugs in progress story points at the end of a sprint
CSV_FIELD_BUG_IN_PROGRESS = 'Bugs in progress story points'
# Bugs not started story points at the end of a sprint
CSV_FIELD_BUG_NOT_STARTED = 'Bugs not started story points'
# Total number of bugs / total number of issues
CSV_FIELD_DEFECT_DENSITY_RATIO = 'Defect density ratio'
# Resolved points / planned points ratio
CSV_FIELD_PREDICTABILITY_RATIO = 'Predictability ratio'
# Total story points / planned story points * 100
CSV_FIELD_SCOPE_CHANGE_PERCENTAGE = 'Scope change percentage'
# Goals achieved / goals planned * 100
CSV_FIELD_GOAL_COMPLETION_PERCENTAGE = 'Goal completion percentage'
# Working days in a sprint
CSV_FIELD_WORKING_DAYS = 'Working days'
# Planned story points / working days
CSV_FIELD_PLANNED_SP_PER_DAY = 'Planned story points per working day'
# Total story points / working days
CSV_FIELD_TOTAL_SP_PER_DAY = 'Total story points per working day'
# Resolved story points / working days
CSV_FIELD_RESOLVED_SP_PER_DAY = 'Velocity per working day'
# Proportion of time work was actively progressing versus waiting. resolved / (resolved + in progress + in code review + in qe) * 100
CSV_FIELD_FLOW_EFFICIENCY = 'Flow efficiency'
# Proportion of time work was actively progressing versus waiting. resolved bugs / (resolved bugs + in progress bugs + in code review bugs + in qe bugs) * 100
CSV_FIELD_FLOW_EFFICIENCY_BUGS = 'Bug flow efficiency'
# Proportion of work that remains incomplete at the end of the sprint. (in progress + in code review + in qe + not started) / planned * 100
CSV_FIELD_WIP_RATIO = 'WIP ratio'
# How drastically the sprint scope changes due to story points being added or removed after the sprint has started
# (added + removed) / planned * 100
CSV_FIELD_SCOPE_VOLATILITY = 'Scope volatility'
# How many of the bugs were resolved vs planned/added. resolved bugs / total bugs * 100
CSV_FIELD_DEFECT_REMOVAL_EFFICIENCY = 'Defect removal efficiency'

CSV_HEADERS = [
    CSV_FIELD_SPRINT_NAME,
    CSV_FIELD_PLANNED_TOTAL,
    CSV_FIELD_RESOLVED_TOTAL,
    CSV_FIELD_NOT_STARTED_TOTAL,
    CSV_FIELD_IN_PROGRESS_TOTAL,
    CSV_FIELD_CODE_REVIEW_TOTAL,
    CSV_FIELD_BUG_TOTAL,
    CSV_FIELD_BUG_RESOLVED,
    CSV_FIELD_BUG_WAITING_QE,
    CSV_FIELD_BUG_WAITING_CODE_REVIEW,
    CSV_FIELD_BUG_IN_PROGRESS,
    CSV_FIELD_BUG_NOT_STARTED,
    CSV_FIELD_WORKING_DAYS,
    CSV_FIELD_PLANNED_SP_PER_DAY,
    CSV_FIELD_TOTAL_SP_PER_DAY,
    CSV_FIELD_RESOLVED_SP_PER_DAY,
    CSV_FIELD_GOAL_COMPLETION_PERCENTAGE,
    CSV_FIELD_SCOPE_CHANGE_PERCENTAGE,
    CSV_FIELD_PREDICTABILITY_RATIO,
    CSV_FIELD_FLOW_EFFICIENCY,
    CSV_FIELD_WIP_RATIO,
    CSV_FIELD_SCOPE_VOLATILITY,
    CSV_FIELD_DEFECT_DENSITY_RATIO,
    CSV_FIELD_DEFECT_REMOVAL_EFFICIENCY,
    CSV_FIELD_FLOW_EFFICIENCY_BUGS,
]

AI_PROMPTS = [
    '''Analyze this data file, corresponding to a development team following scrum practices.
    Summarize these metrics and their values: velocity, velocity per working day, goal completion,
    flow efficiency, wip ratios, defect density ratio and removal efficiency, scope change, scope volatility and predictability.''',
    'Provide charts, including average and trend, for each of the metrics above. Do them separately.',
    'What changes should the team adopt that would have the highest impact? Explain them in start/stop/continue doing.',
]


def sprints_in_board(server, board_id, sprint_list):
    def _filter_sprints(sprint):
        return any(sprint.name == s.name for s in sprint_list)
    all_sprints = server.sprints(board_id, startAt=SKIP_FIRST_N_SPRINTS)
    return list(filter(_filter_sprints, all_sprints))


def issues_from_ushift(server, query):
    issues = []
    i = 0
    chunk_size = 500
    while True:
        chunk = server.search_issues(query, expand='changelog', startAt=i, maxResults=chunk_size)
        i += chunk_size
        issues += chunk.iterable
        if i >= chunk.total:
            break
    return issues


def issue_planned_in_sprint(sprint_id, sprint_name, start_date):
    def _issue_planned_in_sprint(issue):
        def _filter_changes(change):
            return change.created <= start_date and \
                   any(item.field == 'Sprint' for item in change.items)
        filtered_changes = list(filter(_filter_changes, issue.changelog.histories))
        if len(filtered_changes) == 0:
            # An empty filtered_changes could happen either because the sprint was configured when creating the ticket
            # or it was added to the sprint after the start date.
            if any(
                 item.field == 'Sprint' and
                 str(sprint_id) in item.to and
                 (item.fromString is None or sprint_name not in item.fromString) and
                 change.created >= start_date
                 for change in issue.changelog.histories for item in change.items):
                return False
            if issue.fields.created >= start_date:
                return False
            if issue.fields.customfield_12310940 is not None and \
               str(sprint_id) in "".join(issue.fields.customfield_12310940):
                return True
            return False
        return any(item.to is not None and str(sprint_id) in item.to for item in filtered_changes[-1].items)
    return _issue_planned_in_sprint


def issue_removed_from_sprint(sprint_name, start_date, end_date):
    def _issue_removed_from_sprint(issue):
        def _filter_sprint(change):
            return change.created <= end_date and change.created >= start_date and \
                any(item.field == 'Sprint' for item in change.items)

        filtered_list = list(filter(_filter_sprint, issue.changelog.histories))
        if len(filtered_list) == 0:
            return False
        return issue_added_to_sprint(issue, sprint_name, end_date) and \
            any(sprint_name not in item.toString for item in filtered_list[-1].items)

    return _issue_removed_from_sprint


def issue_added_to_sprint(issue, sprint_name, end_date):
    if any(
         item.field == 'Sprint' and
         sprint_name in item.toString and
         change.created <= end_date
         for change in issue.changelog.histories
         for item in change.items):
        return True
    return issue_in_sprint(sprint_name, end_date)(issue)


def issue_in_sprint(sprint_name, end_date):
    def _issue_in_sprint(issue):
        def _filter_changes(change):
            return change.created <= end_date and any(item.field == 'Sprint' for item in change.items)
        filtered_changes = list(filter(_filter_changes, issue.changelog.histories))
        if len(filtered_changes) == 0:
            return issue.fields.customfield_12310940 is not None and \
                   sprint_name in "".join(issue.fields.customfield_12310940)
        return any(item.field == 'Sprint' and item.toString is not None and sprint_name in item.toString for item in filtered_changes[-1].items)
    return _issue_in_sprint


def issue_get_final_state_before_date(issue, end_date):
    def _filter_state_changes(change):
        return change.created <= end_date and \
               any(item.field == 'status' for item in change.items)
    filtered_changes = list(filter(_filter_state_changes, issue.changelog.histories))
    if len(filtered_changes) == 0:
        # if there are no entries default to this state
        return "To Do"
    return filtered_changes[-1].items[-1].toString


def issue_resolved(end_date):
    def _issue_resolved(issue):
        def _filter_resolution(change):
            return change.created <= end_date and \
                   any(item.field == 'resolution'
                       for item in change.items)

        if issue_get_final_state_before_date(issue, end_date) in ['Verified', 'VERIFIED', 'Closed', 'CLOSED']:
            return True
        filtered_changes = list(filter(_filter_resolution, issue.changelog.histories))
        if len(filtered_changes) == 0:
            return False
        return any(item.field == 'resolution' and item.toString is not None and 'Done' in item.toString for item in filtered_changes[-1].items)

    return _issue_resolved


def issue_not_started(end_date):
    def _issue_not_started(issue):
        return issue_get_final_state_before_date(issue, end_date) in ['To Do', 'ToDo', 'New']
    return _issue_not_started


def issue_in_progress(end_date):
    def _issue_in_progress(issue):
        return issue_get_final_state_before_date(issue, end_date) in ['ASSIGNED', 'In Progress']
    return _issue_in_progress


def issue_in_code_review(end_date):
    def _issue_in_code_review(issue):
        return issue_get_final_state_before_date(issue, end_date) in ['Code Review', 'POST']
    return _issue_in_code_review


def issue_in_review(end_date):
    def _issue_in_review(issue):
        return issue_get_final_state_before_date(issue, end_date) in ['Review', 'MODIFIED', 'ON_QA']
    return _issue_in_review


def issue_is_bug(issue):
    return issue.fields.issuetype.name == 'Bug'


def estimated_story_points(issues, end_date):
    def _filter_changes(change):
        return any(item.field == 'Story Points' and
                   len(item.toString) > 0 and
                   change.created <= end_date
                   for item in change.items)

    def _get_story_points(issue):
        filtered_changes = list(filter(_filter_changes, issue.changelog.histories))
        if len(filtered_changes) == 0:
            return int(issue.fields.customfield_12310243) if issue.fields.customfield_12310243 is not None else 0
        return sum(int(item.toString) if item.field == 'Story Points' else 0 for item in filtered_changes[-1].items)

    return sum(_get_story_points(issue) for issue in issues)


def prepare_csv_row(
    sprint_name,
    end_date,
    planned_issues,
    removed_issues,
    total_sprint_issues,
    resolved_issues,
    not_started_issues,
    in_progress_issues,
    in_code_review_issues,
    all_bug_issues,
    resolved_bug_issues,
    in_qa_bug_issues,
    in_code_review_bug_issues,
    in_progress_bug_issues,
    not_started_bug_isses,
    config
):
    goals_achieved, goals_total, working_days = 0, 1, 1
    for sprint in config.sprints:
        if sprint.name == sprint_name:
            goals_achieved = sprint.goals_achieved
            goals_total = sprint.goals_total
            working_days = sprint.working_days
            break

    name = int(''.join(filter(str.isdigit, sprint_name)))
    planned_story_points = estimated_story_points(planned_issues, end_date)
    removed_story_points = estimated_story_points(removed_issues, end_date)
    total_story_points = estimated_story_points(total_sprint_issues, end_date)
    resolved_story_points = estimated_story_points(resolved_issues, end_date)
    not_started_story_points = estimated_story_points(not_started_issues, end_date)
    in_progress_story_points = estimated_story_points(in_progress_issues, end_date)
    in_code_review_story_points = estimated_story_points(in_code_review_issues, end_date)
    all_bug_story_points = estimated_story_points(all_bug_issues, end_date)
    resolved_bug_story_points = estimated_story_points(resolved_bug_issues, end_date)
    in_qa_bug_story_points = estimated_story_points(in_qa_bug_issues, end_date)
    in_code_review_bug_story_points = estimated_story_points(in_code_review_bug_issues, end_date)
    in_progress_bug_story_points = estimated_story_points(in_progress_bug_issues, end_date)
    not_started_bug_story_points = estimated_story_points(not_started_bug_isses, end_date)

    return {
        CSV_FIELD_SPRINT_NAME: name,
        CSV_FIELD_PLANNED_TOTAL: planned_story_points,
        CSV_FIELD_RESOLVED_TOTAL: resolved_story_points,
        CSV_FIELD_NOT_STARTED_TOTAL: not_started_story_points,
        CSV_FIELD_IN_PROGRESS_TOTAL: in_progress_story_points,
        CSV_FIELD_CODE_REVIEW_TOTAL: in_code_review_story_points,
        CSV_FIELD_BUG_TOTAL: all_bug_story_points,
        CSV_FIELD_BUG_RESOLVED: resolved_bug_story_points,
        CSV_FIELD_BUG_WAITING_QE: in_qa_bug_story_points,
        CSV_FIELD_BUG_WAITING_CODE_REVIEW: in_code_review_bug_story_points,
        CSV_FIELD_BUG_IN_PROGRESS: in_progress_bug_story_points,
        CSV_FIELD_BUG_NOT_STARTED: not_started_bug_story_points,
        CSV_FIELD_DEFECT_DENSITY_RATIO: round(len(all_bug_issues) / len(total_sprint_issues), 2),
        CSV_FIELD_PREDICTABILITY_RATIO: round(resolved_story_points / planned_story_points, 2),
        CSV_FIELD_SCOPE_CHANGE_PERCENTAGE: round((total_story_points - planned_story_points) / planned_story_points * 100, 2),
        CSV_FIELD_GOAL_COMPLETION_PERCENTAGE: round(float(goals_achieved/goals_total)*100, 2),
        CSV_FIELD_WORKING_DAYS: working_days,
        CSV_FIELD_PLANNED_SP_PER_DAY: round(planned_story_points / working_days, 2),
        CSV_FIELD_TOTAL_SP_PER_DAY: round(total_story_points / working_days, 2),
        CSV_FIELD_RESOLVED_SP_PER_DAY: round(resolved_story_points / working_days, 2),
        CSV_FIELD_FLOW_EFFICIENCY: round(resolved_story_points / (resolved_story_points + in_progress_story_points + in_code_review_story_points + in_qa_bug_story_points) * 100, 2),
        CSV_FIELD_WIP_RATIO: round((in_progress_story_points + in_code_review_story_points + in_qa_bug_story_points + not_started_story_points) / planned_story_points, 2),
        CSV_FIELD_SCOPE_VOLATILITY: round((total_story_points - planned_story_points + removed_story_points) / planned_story_points * 100, 2),
        CSV_FIELD_DEFECT_REMOVAL_EFFICIENCY: round(resolved_bug_story_points / all_bug_story_points * 100, 2),
        CSV_FIELD_FLOW_EFFICIENCY_BUGS: round(
            resolved_bug_story_points / (resolved_bug_story_points + in_progress_bug_story_points + in_code_review_bug_story_points + in_qa_bug_story_points) * 100, 2),
    }


def write_results_csv(path, rows):
    with open(path, 'w', newline='') as csvfile:
        writer = csv.DictWriter(csvfile, fieldnames=CSV_HEADERS)
        writer.writeheader()
        for row in rows:
            writer.writerow(row)
    logging.info(f"Wrote CSV to {path}")


class Sprint(object):
    def __init__(self, json_dict) -> None:
        mandatory_fields = ['name', 'goals_total', 'goals_achieved', 'working_days']
        for field in mandatory_fields:
            if field not in json_dict:
                raise ValueError(f'{field} missing from sprints configuration')
            setattr(self, field, json_dict[field])


class Config(object):
    def __init__(self, json_dict) -> None:
        mandatory_fields = ['board_id', 'jira_query', 'jira_token', 'results_path', 'raw_issues_path']
        for field in mandatory_fields:
            if field not in json_dict:
                raise ValueError(f'{field} missing from configuration')
            setattr(self, field, json_dict[field])
        self.sprints = []
        for sprint in json_dict['sprints']:
            self.sprints.append(Sprint(sprint))


def load_sprint_config(path):
    with open(path, 'r') as fd:
        config = json.load(fd)
        return Config(config)


def write_raw(path, sprints, issues):
    with open(path, 'w') as fd:
        for issue in issues:
            fd.write(json.dumps(issue.raw))
            fd.write('\n')
        for sprint in sprints:
            fd.write(json.dumps(sprint.raw))
            fd.write('\n')


def main(args):
    logging.basicConfig(level=logging.INFO, format="{asctime} - {levelname} - {message}", style="{", datefmt="%Y-%m-%d %H:%M:%S")

    config = load_sprint_config(args.config)

    server = JIRA(server='https://issues.redhat.com', token_auth=config.jira_token)
    logging.info("Logged into JIRA")

    sprints = sprints_in_board(server, config.board_id, config.sprints)
    logging.info(f"Retrieved {len(sprints)} sprints")

    all_issues = issues_from_ushift(server, config.jira_query)
    logging.info(f"Retrieved {len(all_issues)} issues")

    csv_rows = []
    for sprint in sprints:
        logging.info(f"Sprint {sprint.name}. ID {sprint.id}. Start {sprint.startDate}. End {sprint.endDate}")
        if sprint.state != 'closed':
            logging.info("Sprint has not finished yet")
            continue

        start_date, end_date = sprint.activatedDate, sprint.completeDate

        planned_issues = list(filter(issue_planned_in_sprint(sprint.id, sprint.name, start_date), all_issues))
        removed_issues = list(filter(issue_removed_from_sprint(sprint.name, start_date, end_date), all_issues))
        total_sprint_issues = list(filter(issue_in_sprint(sprint.name, end_date), all_issues))
        resolved_issues = list(filter(issue_resolved(end_date), total_sprint_issues))
        not_started_issues = list(filter(issue_not_started(end_date), total_sprint_issues))
        in_progress_issues = list(filter(issue_in_progress(end_date), total_sprint_issues))
        in_code_review_issues = list(filter(issue_in_code_review(end_date), total_sprint_issues))
        all_bug_issues = list(filter(issue_is_bug, total_sprint_issues))
        resolved_bug_issues = list(filter(issue_resolved(end_date), all_bug_issues))
        in_qa_bug_issues = list(filter(issue_in_review(end_date), all_bug_issues))
        in_code_review_bug_issues = list(filter(issue_in_code_review(end_date), all_bug_issues))
        in_progress_bug_issues = list(filter(issue_in_progress(end_date), all_bug_issues))
        not_started_bug_issues = list(filter(issue_not_started(end_date), all_bug_issues))

        csv_row = prepare_csv_row(
            sprint.name,
            end_date,
            planned_issues,
            removed_issues,
            total_sprint_issues,
            resolved_issues,
            not_started_issues,
            in_progress_issues,
            in_code_review_issues,
            all_bug_issues,
            resolved_bug_issues,
            in_qa_bug_issues,
            in_code_review_bug_issues,
            in_progress_bug_issues,
            not_started_bug_issues,
            config,
        )
        logging.info(f"Finished processing. {csv_row}")

        csv_rows.append(csv_row)

    write_results_csv(config.results_path, csv_rows)

    write_raw(config.raw_issues_path, sprints, all_issues)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(
        description=__doc__,
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    parser.add_argument(
        '--config',
        default='sprint_metrics_config.json',
        help='Configuration file path',
    )
    args = parser.parse_args()
    main(args)
