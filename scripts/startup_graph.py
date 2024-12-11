import json
import pandas as pd
import plotly.express as px
import datetime

def format_time_ns(nanoseconds):
    """
    Convert time from nanoseconds to a human-readable string.

    Args:
    nanoseconds (int): Time in nanoseconds.

    Returns:
    str: Time formatted as a human-readable string (e.g., '1.57 s', '2.89 ms', '123 ns').
    """
    if nanoseconds >= 1e9:  # More than or equal to 1 second
        return f"{nanoseconds / 1e9:.2f} s"
    elif nanoseconds >= 1e6:  # More than or equal to 1 millisecond
        return f"{nanoseconds / 1e6:.2f} ms"
    elif nanoseconds >= 1e3:  # More than or equal to 1 microsecond
        return f"{nanoseconds / 1e3:.2f} μs"
    else:  # Less than 1 microsecond
        return f"{nanoseconds} ns"

with open("scripts/startup_times.json", "r") as f:
    json_data = json.load(f)

services = json_data["services"]
microshift_start = json_data["microshiftStart"]
microshift_ready = json_data["microshiftReady"]

df = pd.DataFrame(services)

df["start"] = pd.to_datetime(df["start"])
df["ready"] = pd.to_datetime(df["ready"])
df["timeToReadyReadable"] = df["timeToReady"].apply(format_time_ns)

#print(df.dtypes)

fig = px.timeline(
    df,
    x_start="start",
    x_end="ready",
    y="name",
    color="timeToReady",
    hover_data={
        "timeToReady": False,
        "timeToReadyReadable": True,
        "dependencies": True
    },
    color_continuous_scale="RdYlGn_r"  # Use reversed Red-Yellow-Green scale for green-to-red
)
fig.update_yaxes(autorange="reversed")
fig.update_layout(coloraxis_showscale=False)


# Convert numerical x-coordinates to datetime
def draw_dep(fig, serv1, serv2):
    num1 = df.loc[df["name"] == serv1].index[0]
    num2 = df.loc[df["name"] == serv2].index[0]


    x0 = fig.data[0].base[num1] + datetime.timedelta(milliseconds=(fig.data[0].x[num1]))
    x1 = fig.data[0].base[num2] #+ datetime.timedelta(seconds=(fig.data[0].x[6]/1000000000))

    # Add line shape using the corrected datetime values
    fig.add_shape(
        type="line",
        x0=x0,
        x1=x1,
        y0=fig.data[0].y[num1],  # Assuming y is correct
        y1=fig.data[0].y[num2],
        line=dict(color="Red", width=2),
        xref="x",
        yref="y"  # Match the axis types
    )
    return fig

fig = draw_dep(fig, "etcd", "kube-apiserver")


fig.show()
