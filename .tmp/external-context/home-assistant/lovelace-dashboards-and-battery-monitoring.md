---
source: Context7 API + official Home Assistant docs
library: Home Assistant
package: home-assistant
topic: lovelace dashboards api sections view battery monitoring
fetched: 2026-04-20T00:00:00Z
official_docs: https://www.home-assistant.io/dashboards/
---

# Home Assistant Lovelace dashboards + battery monitoring (filtered)

## Creating dashboards

- Officially documented creation path for new dashboards: **Settings > Dashboards > Add dashboard**.
- Home Assistant distinguishes:
  - **Built-in dashboards** (limited editing)
  - **User-created dashboards**
  - **YAML dashboards** under `lovelace:` in `configuration.yaml`
- Official docs say **storage-mode dashboards can be created in the configuration panel**; YAML dashboards are defined in config files.

### YAML dashboard definition

```yaml
lovelace:
  resource_mode: yaml
  resources:
    - url: /local/my-custom-card.js
      type: module
  dashboards:
    battery-monitor:
      mode: yaml
      filename: battery-monitor.yaml
      title: Battery Monitor
      icon: mdi:battery-heart
      show_in_sidebar: true
```

Minimal dashboard YAML:

```yaml
views:
  - title: Batteries
    cards:
      - type: markdown
        content: >
          Battery overview
```

## Modern recommended dashboard patterns

- Current docs describe **Sections** as the default view type.
- Current docs/blog recommend **Tile card** as the primary suggested card for Sections view.
- The built-in **Home dashboard** uses **Sections view + Tile cards**.
- Sections view organizes cards in sections on a grid and supports:
  - section headings
  - badges in the header
  - drag-and-drop rearrangement
  - conditional section visibility
  - section backgrounds
  - responsive layout preserving card grouping

### Sections view YAML highlights

```yaml
type: sections
max_columns: 4   # set in UI; docs expose layout options and section config
```

Relevant section options from docs:

- header supports title card + badges
- section `background: true` or custom `{ color, opacity }`
- footer supports one card pinned at bottom

### Why Sections over Masonry

- The dashboard design docs explain Masonry is harder to predict across screen sizes.
- Sections were introduced for more predictable grouping, responsiveness, and drag/drop layout.
- Cards currently optimized around the grid system include **Tile**, **Button**, and **Sensor** cards.

## Native cards most useful for battery-powered device monitoring

### 1) Tile card

Best for per-device battery status in Sections view.

```yaml
type: tile
entity: sensor.front_door_lock_battery
color: state
state_content:
  - state
  - last_changed
vertical: false
```

Useful options:

- `state_content` can show `state`, `last_changed`, `last_updated`, or attributes
- `color: state` uses entity state/device_class/domain coloring
- works well as the default card in Sections view

### 2) Entity filter card

Useful for a “Low batteries” section that only appears when something is below a threshold.

```yaml
type: entity-filter
entities:
  - sensor.front_door_lock_battery
  - sensor.kitchen_motion_battery
  - sensor.window_sensor_battery
conditions:
  - condition: numeric_state
    below: 20
show_empty: false
```

Notes:

- YAML-only card
- can render as default Entities card or another card such as Glance
- supports numeric thresholds and hiding itself when empty

### 3) Area card

Useful if your battery-powered devices are well-organized by area and you want room-level drill-down.

```yaml
type: area
area: hallway
tap_action:
  action: navigate
  navigation_path: /battery-monitor/hallway
```

### 4) Statistic card

Useful if you build derived/statistical battery entities and want summaries over time.

```yaml
type: statistic
entity: sensor.front_door_lock_battery
stat_type: min
period:
  calendar:
    period: month
```

## Native entity / data patterns for battery monitoring

- Official battery tutorial assumes battery percentages are usually reported as **`sensor` entities with `device_class: battery`**.
- Docs recommend checking **Developer Tools > States** and filtering on `battery`.
- Modern template pattern for low batteries:
  - iterate `states.sensor`
  - filter `attributes.device_class == 'battery'`
  - reject `unknown` / `unavailable`
  - compare numeric state against threshold
  - use `device_name()` instead of sensor name
  - use `area_name()` for room context

### Current official low-battery template

```jinja
{% set low = namespace(batteries=[]) %}
{% for sensor in states.sensor
   | selectattr('attributes.device_class', 'eq', 'battery')
   | rejectattr('state', 'in', ['unknown', 'unavailable']) %}
  {% if sensor.state | float(100) < 20 %}
    {% set device = device_name(sensor.entity_id) %}
    {% set area = area_name(sensor.entity_id) %}
    {% set label = device ~ (' in ' ~ area if area else '')
       ~ ' (' ~ sensor.state ~ '%)' %}
    {% set low.batteries = low.batteries + [label] %}
  {% endif %}
{% endfor %}
{% if low.batteries | count == 0 %}
  All batteries are healthy.
{% elif low.batteries | count == 1 %}
  1 device needs a new battery: {{ low.batteries[0] }}.
{% else %}
  {{ low.batteries | count }} devices need new batteries: {{ low.batteries | join(', ') }}.
{% endif %}
```

## Practical battery dashboard pattern

Recommended structure from current docs/patterns:

1. Create a **new user dashboard** instead of heavily modifying the built-in one.
2. Use **Sections view**.
3. Add sections such as:
   - **Low batteries** → `entity-filter`
   - **All battery devices** → Tile cards grouped by area/device type
   - **By room** → Area cards for navigation
   - **Trends / worst performers** → Statistic cards where useful
4. Ensure devices/entities are properly assigned to **areas**.
5. Prefer **device names + area names** in templates/alerts because many battery sensors are generically named “Battery”.

## Sources used

- https://www.home-assistant.io/dashboards/
- https://www.home-assistant.io/dashboards/dashboards/
- https://www.home-assistant.io/dashboards/sections/
- https://www.home-assistant.io/dashboards/tile/
- https://www.home-assistant.io/dashboards/entity-filter/
- https://www.home-assistant.io/dashboards/area/
- https://www.home-assistant.io/dashboards/statistic/
- https://www.home-assistant.io/docs/templating/tutorial-battery-alerts/
- https://www.home-assistant.io/blog/2024/03/04/dashboard-chapter-1/
