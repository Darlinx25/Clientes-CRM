-- Revierte el renombrado de "RtasFinExter" a "Rentas Exterior".
UPDATE company_types SET name = 'Rentas Exterior', updated_at = datetime('now')
    WHERE name = 'RtasFinExter';