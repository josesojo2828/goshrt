#!/usr/bin/env bash
set -euo pipefail

API="${GOSHRT_API:-http://localhost:15050}"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# ─── helpers ───────────────────────────────────────────────────────────

json() { python3 -m json.tool 2>/dev/null || cat; }

ok()   { printf " ${GREEN}✅${NC} %s\n" "$1"; }
err()  { printf " ${RED}❌${NC} %s\n" "$1"; }
info() { printf " ${CYAN}ℹ️${NC} %s\n" "$1"; }
title(){ printf "\n${BOLD}${CYAN}═══ %s ═══${NC}\n\n" "$1"; }

# ─── non-interactive (batch mode) ──────────────────────────────────────

batch_add() {
	local url alias ttl
	url="${1:-}"; alias=""; ttl=""
	shift 2>/dev/null || true
	while [[ $# -gt 0 ]]; do
		case "$1" in
			--alias|-a) alias="$2"; shift 2 ;;
			--ttl|-t)   ttl="$2";   shift 2 ;;
			*) err "Unknown: $1"; exit 1 ;;
		esac
	done
	[[ -z "$url" ]] && { err "Usage: $0 add <url> [--alias <slug>] [--ttl <sec>]"; exit 1; }

	body=$(jq -n --arg u "$url" '{"url":$u}')
	[[ -n "$alias" ]] && body=$(echo "$body" | jq --arg a "$alias" '. + {custom_alias:$a}')
	[[ -n "$ttl" ]]   && body=$(echo "$body" | jq --arg t "$ttl" '. + {ttl_seconds:($t|tonumber)}')

	curl -s -X POST "$API/api/url" -H "Content-Type: application/json" -d "$body" | json
}

batch_list() { curl -s "$API/api/urls" | json; }

batch_stats() { curl -s "$API/api/url/$1/stats" | json; }

batch_delete() { curl -s -X DELETE "$API/api/url/$1" | json; }

batch_open() {
	printf " ${CYAN}🔓${NC} Following http://localhost:15050/%s ...\n" "$1"
	local final_url
	final_url=$(curl -sL -o /dev/null -w '%{url_effective}' "http://localhost:15050/$1")
	printf " → ${GREEN}%s${NC}\n" "$final_url"
}

# ─── interactive mode ──────────────────────────────────────────────────

pause() { printf "\n ${YELLOW}Presioná Enter para volver al menú...${NC}" && read -r; }

create_url() {
	title "CREAR URL CORTA"
	read -r -p "  URL original: " url
	[[ -z "$url" ]] && { err "La URL es obligatoria"; pause; return; }

	read -r -p "  Slug personalizado (Enter = aleatorio): " alias

	read -r -p "  TTL en segundos (Enter = sin expiración): " ttl_input

	body=$(jq -n --arg u "$url" '{"url":$u}')
	[[ -n "$alias" ]] && body=$(echo "$body" | jq --arg a "$alias" '. + {custom_alias:$a}')
	[[ -n "$ttl_input" ]] && body=$(echo "$body" | jq --arg t "$ttl_input" '. + {ttl_seconds:($t|tonumber)}')

	echo ""
	result=$(curl -s -X POST "$API/api/url" -H "Content-Type: application/json" -d "$body")

	if echo "$result" | jq -e '.short_code' >/dev/null 2>&1; then
		code=$(echo "$result" | jq -r '.short_code')
		ok "Creada! → ${BOLD}$code${NC}  →  localhost:15050/$code"
	else
		err "$(echo "$result" | jq -r '.error // "error desconocido"')"
	fi
	pause
}

list_urls() {
	title "LISTA DE URLs"
	result=$(curl -s "$API/api/urls")
	urls=$(echo "$result" | jq -c '.urls[]' 2>/dev/null) || {
		err "No se pudieron obtener las URLs"; pause; return
	}

	if [[ -z "$urls" ]]; then
		info "No hay URLs todavía."
	else
		printf "  ${BOLD}%-10s %-25s %-20s ${NC}\n" "SLUG" "DESTINO" "CREADO"
		printf "  %s\n" "$(printf '─%.0s' {1..65})"
		while IFS= read -r u; do
			code=$(echo "$u" | jq -r '.short_code')
			orig=$(echo "$u" | jq -r '.original_url')
			created=$(echo "$u" | jq -r '.created_at' | cut -dT -f1)
			printf "  %-10s %-25s %-20s\n" "$code" "$orig" "$created"
		done <<< "$urls"
	fi
	echo ""
	info "Total: $(echo "$result" | jq -r '.urls | length') URLs"
	pause
}

stats_url() {
	title "ESTADÍSTICAS"
	read -r -p "  Slug de la URL: " code
	[[ -z "$code" ]] && { err "Slug requerido"; pause; return; }

	result=$(curl -s "$API/api/url/$code/stats")
	if echo "$result" | jq -e '.short_code' >/dev/null 2>&1; then
		echo ""
		echo "$result" | jq -r '
			"  Slug:        \(.short_code)
			  URL:         \(.original_url)
			  Clicks:      \(.clicks)
			  Creado:      \(.created_at)
			  Último acc:  \(.last_accessed // "nunca")
			  Activo:      \(.is_active)
			  Expira:      \(.expires_at // "no expira")"
		'
	else
		err "$(echo "$result" | jq -r '.error // "slug no encontrado"')"
	fi
	pause
}

delete_url() {
	title "ELIMINAR URL"
	read -r -p "  Slug a eliminar: " code
	[[ -z "$code" ]] && { err "Slug requerido"; pause; return; }

	# confirmar
	info "¿Estás seguro de eliminar '${BOLD}$code${NC}'?"
	read -r -p "  Escribí el slug de nuevo para confirmar: " confirm
	[[ "$confirm" != "$code" ]] && { err "No coincide, cancelado."; pause; return; }

	result=$(curl -s -X DELETE "$API/api/url/$code")
	if echo "$result" | jq -e '.message' >/dev/null 2>&1; then
		ok "Eliminada!"
	else
		err "$(echo "$result" | jq -r '.error // "error"')"
	fi
	pause
}

open_url() {
	title "ABRIR URL"
	read -r -p "  Slug: " code
	[[ -z "$code" ]] && { err "Slug requerido"; pause; return; }

	url=$(curl -s "$API/api/url/$code/stats" | jq -r '.original_url // empty')
	if [[ -n "$url" ]]; then
		ok "Redirige a: ${BOLD}$url${NC}"
		curl -sL -o /dev/null -w "  Status: %{http_code} → ${GREEN}%{url_effective}${NC}\n" "http://localhost:15050/$code"
	else
		err "Slug no encontrado"
	fi
	pause
}

# ─── show_menu ─────────────────────────────────────────────────────────

show_menu() {
	clear
	printf '%s' "$CYAN"
	printf '   ╔═══════════════════════════════╗\n'
	printf '   ║       goshrt — CLI 🚀         ║\n'
	printf '   ╚═══════════════════════════════╝\n'
	printf '%s' "$NC"
	printf "   ${BOLD}API:${NC} %s\n\n" "$API"
	printf "   ${BOLD}1.${NC}  Crear URL corta\n"
	printf "   ${BOLD}2.${NC}  Listar URLs\n"
	printf "   ${BOLD}3.${NC}  Estadísticas\n"
	printf "   ${BOLD}4.${NC}  Abrir URL (follow redirect)\n"
	printf "   ${BOLD}5.${NC}  Eliminar URL\n"
	printf "   ${BOLD}0.${NC}  Salir\n\n"
}

interactive() {
	while true; do
		show_menu
		read -r -p "   Elegí una opción [0-5]: " opt
		case "$opt" in
			1) create_url ;;
			2) list_urls ;;
			3) stats_url ;;
			4) open_url ;;
			5) delete_url ;;
			0) echo -e "\n ${GREEN}👋${NC} Chau!\n"; exit 0 ;;
			*) err "Opción inválida"; sleep 1 ;;
		esac
	done
}

# ─── entrypoint ────────────────────────────────────────────────────────

if [[ $# -eq 0 ]]; then
	interactive
	exit 0
fi

cmd="$1"; shift
case "$cmd" in
	add|a)     batch_add "$@" ;;
	list|l)    batch_list ;;
	stats|s)   batch_stats "${1:-}";;
	delete|d)  batch_delete "${1:-}";;
	open|o)    batch_open "${1:-}";;
	help|h|*)  echo "Usage: $0 [add|list|stats|delete|open|help]"; exit 0 ;;
esac
