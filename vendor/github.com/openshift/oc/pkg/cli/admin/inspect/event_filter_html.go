package inspect

const eventHTMLPage = `
<!doctype html>
<html lang="en">
  <head>
    <!-- Required meta tags -->
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">

    <!-- Bootstrap CSS -->
    <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css" integrity="sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T" crossorigin="anonymous">
    <link rel="stylesheet" href="https://use.fontawesome.com/releases/v5.6.3/css/all.css" integrity="sha384-UHRtZLI+pbxtHCWp1t77Bi1L4ZtiqrqD80Kn4Z8NTSRyMA2Fd33n5dQ8lWUE00s/" crossorigin="anonymous">
	<link href="https://unpkg.com/bootstrap-table@1.17.1/dist/bootstrap-table.min.css" rel="stylesheet">

    <title>Events</title>
	<style type="text/css">
      body * {
       font-size: 12px!important;
      }
	  .text-overflow() {
       overflow: hidden;
	   text-overflow: ellipsis;
	   white-space: nowrap;
	  }
      .truncated {
        display: inline-block;
        max-width: 200px;
        .text-overflow();
      }
    </style>
  </head>
  <body>

<table
  id="events"
  class="table table-bordered table-hover table-sm"
  data-toggle="table"
  data-search="true"
  data-show-search-clear-button="true"
  data-filter-control="true"
  data-advanced-search="true"
  data-id-table="advancedTable"
  data-pagination="true"
  data-page-size="100"
  data-show-columns-toggle-all="true"
  data-show-pagination-switch="true"
  data-show-columns="true">
  <thead>
    <tr>
      <th data-width="100" data-field="time" data-filter-control="input" data-sortable="true">Time</th>
      <th data-width="200" data-field="namespace" data-filter-control="input" data-sortable="true">Namespace</th>
      <th data-width="200" data-field="component" data-filter-control="input" data-sortable="true">Component</th>
      <th data-width="200" data-field="relatedobject" data-filter-control="input" data-sortable="true">RelatedObject</th>
      <th data-field="reason" data-filter-control="input">Reason</th>
      <th data-field="message" data-filter-control="input" data-escape="true">Message</th>
    </tr>
  </thead>
  <tbody>
    {{range .Items}}
    <tr>
      <td>{{formatTime .ObjectMeta.CreationTimestamp .FirstTimestamp .LastTimestamp .Count}}</td>
      <td><p class="truncated">{{.Namespace}}</p></td>
      <td><p class="truncated">{{.Source.Component}}</p></td>
      <td><p class="truncated">{{.InvolvedObject.Name}}</p></td>
      <td>{{formatReason .Reason}}</td>
      <td data-formatter="messageFormatter">{{.Message}}</td>
    </tr>
    {{end}}
  </tbody>
</table>

    <!-- Optional JavaScript -->
    <!-- jQuery first, then Popper.js, then Bootstrap JS -->
    <script src="https://code.jquery.com/jquery-3.3.1.min.js" integrity="sha256-FgpCb/KJQlLNfOu91ta32o/NMZxltwRo8QtmkMRdAu8=" crossorigin="anonymous"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.14.7/umd/popper.min.js" integrity="sha384-UO2eT0CpHqdSJQ6hJty5KVphtPhzWj9WO1clHTMGa3JDZwrnQq4sF86dIHNDz0W1" crossorigin="anonymous"></script>
    <script src="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/js/bootstrap.min.js" integrity="sha384-JjSmVgyd0p3pXB1rRibZUAYoIIy6OrQ6VrjIEaFf/nJGzIxFDsf4x0xIM+B07jRM" crossorigin="anonymous"></script>

    <script src="https://unpkg.com/bootstrap-table@1.17.1/dist/bootstrap-table.min.js" crossorigin="anonymous"></script>
	<script src="https://unpkg.com/bootstrap-table@1.17.1/dist/extensions/toolbar/bootstrap-table-toolbar.min.js" crossorigin="anonymous"></script>
    <script src="https://unpkg.com/bootstrap-table@1.17.1/dist/extensions/filter-control/bootstrap-table-filter-control.min.js" crossorigin="anonymous"></script>

	<script>
	function messageFormatter(value, row) {
    	return '<code>'+value+'</code>'
  	}
    $(function() {
      $('#events').bootstrapTable()
    })
	</script>
  </body>
</html>
`
