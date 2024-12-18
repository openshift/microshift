import json
import pandas as pd
import plotly.express as px
import numpy as np

def readable_time(nanoseconds):
    if nanoseconds >= 1e9:
        return f"{nanoseconds / 1e9:.2f} s"
    elif nanoseconds >= 1e6:
        return f"{nanoseconds / 1e6:.2f} ms"
    elif nanoseconds >= 1e3:
        return f"{nanoseconds / 1e3:.2f} μs"
    else:
        return f"{nanoseconds} ns"


with open("scripts/startup_times.json", "r") as f:
    json_data = json.load(f)

services = json_data["services"]
microshift = json_data["microshift"]
microshift_start = pd.to_datetime(microshift["start"])
microshift_serv_start = pd.to_datetime(microshift["servicesStart"])
microshift_ready = pd.to_datetime(microshift["ready"])

df = pd.DataFrame(services)

df["start"] = pd.to_datetime(df["start"])
df["ready"] = pd.to_datetime(df["ready"])
df["adjustedReady"] = df["ready"]
df["timeToReady_ns"] = df["timeToReady"]
df["timeToReady"] = df["timeToReady"].apply(readable_time)

min_width = pd.Timedelta(milliseconds=30)

df["adjustedReady"] = np.where(
    (df["ready"] - df["start"]) < min_width,
    df["start"] + min_width,
    df["ready"]
)

custom_color_scale = [
    (0.0, "#2E7F18"),
    (0.4, "#45731E"),
    (0.6, "#675E24"),
    (0.7, "#8D472B"),
    (0.9, "#B13433"),
    (1.0, "#C82538"),
]

fig = px.timeline(
    df,
    x_start="start",
    x_end="adjustedReady",
    y="name",
    color="timeToReady_ns",
    hover_data={
        "start": False,
        "ready": False,
        "timeToReady": True,
        "timeToReady_ns": True,
        "dependencies": True
    },
    color_continuous_scale=custom_color_scale,
    labels={"name": "Service"} 
)

fig.update_yaxes(autorange="reversed")
fig.update_layout(coloraxis_showscale=False)

fig.update_layout(
    xaxis=dict(type="date", range=[microshift_start - pd.Timedelta(milliseconds=200), microshift_ready + pd.Timedelta(milliseconds=200)]),
)

fig.add_vline(x=microshift_start, line_dash="dash")
#fig.add_vline(x=microshift_serv_start, line_dash="dash")
fig.add_vline(x=microshift_ready, line_dash="dash")

fig.add_annotation(
    x=microshift_start,
    y=1.01,
    yref="paper",
    text="Microshift start",
    showarrow=False,
    xanchor="center",
    yanchor="bottom",
    font=dict(size=12, color="black")
)

fig.add_annotation(
    x=microshift_ready,
    y=1.01,
    yref="paper",
    text="Microshift ready",
    showarrow=False,
    xanchor="center",
    yanchor="bottom",
    font=dict(size=12, color="black")
)

fig.show()
