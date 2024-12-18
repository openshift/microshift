import json
import pandas as pd
import plotly.express as px
import datetime
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
microshift_start = microshift["start"]
microshift_serv_start = microshift["servicesStart"]
microshift_ready = microshift["ready"]

df = pd.DataFrame(services)

df["start"] = pd.to_datetime(df["start"])
df["ready"] = pd.to_datetime(df["ready"])
df["timeToReadyReadable"] = df["timeToReady"].apply(readable_time)

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


def dep_line(x0, y0, x1, y1, x_offset=0, y_offset=0):
    
    x0_1 = x0 + x_offset
    x1_1 = x1 - x_offset

    return [dict(
        type="line",
        x0=x0,
        x1=x0_1,
        y0=y0,  
        y1=y0,
        line=dict(color="Red", width=2),
        xref="x",
        yref="y" 
    ), 
    dict(
        type="line",
        x0=x0_1,
        x1=x1_1,
        y0=y0,  
        y1=y1,
        y1shift=y_offset,
        line=dict(color="Red", width=2),
        xref="x",
        yref="y"
    ),
    dict(
        type="line",
        x0=x1_1,
        x1=x1,
        y0=y1,  
        y1=y1,
        y0shift=y_offset,
        y1shift=y_offset,
        line=dict(color="Red", width=2),
        xref="x",
        yref="y" 
    )]

# Convert numerical x-coordinates to datetime
def draw_dep(fig, serv1, serv2, y_offset=0):
    num1 = df.loc[df["name"] == serv1].index[0]
    num2 = df.loc[df["name"] == serv2].index[0]

    x_offset = datetime.timedelta(milliseconds=200)

    y_offset = y_offset

    x0 = fig.data[0].base[num1] + datetime.timedelta(milliseconds=(fig.data[0].x[num1]))
    y0 = fig.data[0].y[num1]

    x1 = fig.data[0].base[num2]
    y1 = fig.data[0].y[num2]

    shapes = dep_line(x0=x0, y0=y0, x1=x1, y1=y1, x_offset=x_offset, y_offset=y_offset)

    return shapes

    #fig.update_layout(
    #    shapes=shapes
    #)

    # x0_1 = x0 + x_offset
    # x1_1 = x1 - x_offset

    # #Add line shape using the corrected datetime values
    # fig.add_shape(
    #     type="line",
    #     x0=x0,
    #     x1=x0_1,
    #     y0=y0,  
    #     y1=y0,
    #     line=dict(color="Red", width=2),
    #     xref="x",
    #     yref="y"  
    # )

    # fig.add_shape(
    #     type="line",
    #     x0=x0_1,
    #     x1=x1_1,
    #     y0=y0,  
    #     y1=y1,
    #     y1shift=y_offset,
    #     line=dict(color="Red", width=2),
    #     xref="x",
    #     yref="y"  
    # )

    # fig.add_shape(
    #     type="line",
    #     x0=x1_1,
    #     x1=x1,
    #     y0=y1,  
    #     y1=y1,
    #     y0shift=y_offset,
    #     y1shift=y_offset,
    #     line=dict(color="Red", width=2),
    #     xref="x",
    #     yref="y"  
    # )
    
    # x_mid = x0 + (x1-x0)/2
    # print(type(x_mid))
    #y_mid = y1 - y0
    
    #path = f"""
    #M {x0},{y0} C {x_mid},{y0} {x_mid},{y1} {x1},{y1}
    #"""

    # path = f""" M {x0},{y0} L {x1},{y1}"""

    # print(path)
    
    # fig.add_shape(
    #     type="path",
    #     path=path,
    #     line=dict(color="Red", width=2),
    #     #xref="x",
    #     #yref="y"  
    # )

    # fig.add_shape(
    #     type="path",
    #     path="M 2,2 L 5,5",
    #     line=dict(color="Red", width=2),
    #     #xref="x",
    #     #yref="y"  
    # )

    # fig.add_shape(
    #     type="line",
    #     x0=x0,
    #     x1=x_mid,
    #     y0=y0,  
    #     y1=y0,
    #     line=dict(color="Red", width=2),
    #     xref="x",
    #     yref="y"  
    # )

    # fig.add_shape(
    #     type="line",
    #     x0=x_mid,
    #     x1=x_mid,
    #     y0=y0,  
    #     y1=y1,
    #     line=dict(color="Red", width=2),
    #     xref="x",
    #     yref="y"  
    # )

    # fig.add_shape(
    #     type="line",
    #     x0=x_mid,
    #     x1=x1,
    #     y0=y1,  
    #     y1=y1,
    #     line=dict(color="Red", width=2),
    #     xref="x",
    #     yref="y"  
    # )



    #fig.add_annotation(
    #    x=x1,
    #    y=y1,
    #    ax=x0,
    #    ay=y0,
    #    xref="x",
    #    yref="y",
    #    axref="x",
    #    ayref="y",
    #    showarrow=True,
    #    arrowhead=2,
    #    arrowsize=1,
    #    arrowwidth=2,
    #    arrowcolor="Red",
    #    visible=True
    #)
    return fig

#fig = draw_dep(fig, "kube-apiserver", "infrastructure-services-manager")

shapes = []

for service in services:
    for dependency in service["dependencies"]:
        shapes.append(draw_dep(fig, dependency, service["name"]))

new_shapes = []

for shape in shapes:
    for line in shape:
        new_shapes.append(line)

def make_buttons():
    buttons = []
    for service in services:
        deps = []
        num_deps = len(service["dependencies"])
        print(num_deps)
        if (num_deps > 1):
            y_offsets=np.linspace(-0.25, 0.25, num_deps)
        else:
            y_offsets=np.zeros(num_deps)
        for i, dependency in enumerate(reversed(service["dependencies"])):
            deps.extend(draw_dep(fig, dependency, service["name"], y_offsets[i]))
        buttons.append(dict(
            label=service["name"],
            method="relayout",
            args=["shapes", deps],
        ))

    return buttons

buttons = make_buttons()
buttons.insert(0, dict(
    label="None",
    method="relayout",
    args=["shapes", []]
))

fig.update_layout(
    xaxis=dict(type="date", range=[pd.to_datetime(microshift_start), pd.to_datetime(microshift_ready)]),
    updatemenus=[
        dict(
            type="buttons",
            buttons=buttons,
                # dict(
                #     label="None",
                #     method="relayout",
                #     args=["shapes", []]
                # ),
                # dict(
                #     label="All",
                #     method="relayout",
                #     args=["shapes", new_shapes]
                # )
        )
    ]
)

fig.show()
