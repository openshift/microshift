"""Render YAML templates with variable substitution for storage tests."""

import os
import re
import tempfile


class StorageTemplates:

    def render_template(self, template_path, **substitutions):
        """Read a YAML template, replace ${KEY} placeholders, write to a temp file, return its path."""
        with open(template_path) as f:
            content = f.read()
        for key, value in substitutions.items():
            content = content.replace(f"${{{key}}}", str(value))
        unreplaced = re.findall(r"\$\{[A-Z_]+\}", content)
        if unreplaced:
            raise ValueError(
                f"Unreplaced placeholders in {template_path}: {', '.join(unreplaced)}"
            )
        fd, path = tempfile.mkstemp(suffix=".yaml")
        with os.fdopen(fd, "w") as f:
            f.write(content)
        return path

    def cleanup_rendered_file(self, path):
        """Remove a rendered template file."""
        if os.path.exists(path):
            os.unlink(path)
